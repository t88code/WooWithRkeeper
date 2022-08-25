package webhook

import (
	"WooWithRkeeper/internal/bx24api"
	modelsBX24API "WooWithRkeeper/internal/bx24api/models"
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/internal/rk7api"
	modelsRK7API "WooWithRkeeper/internal/rk7api/models"
	"WooWithRkeeper/pkg/logging"
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

func CreateDealInRkeeper(DealID int) error {

	logger := logging.GetLogger()
	logger.Println("Start CreateDealInRkeeper")
	defer logger.Println("End CreateDealInRkeeper")
	cfg := config.GetConfig()

	RK7API := rk7api.NewAPI(cfg.RK7MID.URL, cfg.RK7.User, cfg.RK7.Pass)
	BX24API := bx24api.NewAPI(cfg.BX24.URL)

	logger.Infof("Создана сделка DealID=%d, запущена обработка Webhook на событие создание заказа")
	//получить сделку
	Deal, err := BX24API.DealGet(DealID)
	if err != nil {
		logger.Infof("Не удалось получить инфо по сделке из BX24, error: %v", err)
		return errors.Wrapf(err, "failed BX24API.DealGet(%d)", DealID)
	}
	logger.Info("Детализация по сделке получена успешно")

	//отправить CreateOrder
	//заполнить Order
	Order := new(modelsRK7API.OrderInRK7QueryCreateOrder)
	Order.OrderType.Code = cfg.RK7MID.OrderTypeCode
	Order.Table.Code = cfg.RK7MID.TableCode
	Order.PersistentComment = Deal.COMMENTS
	Order.ExtSource = "BX24"
	Order.ExtID = fmt.Sprint(DealID)

	//заполнить Props
	var Props []*modelsRK7API.Prop
	Props = append(Props, &modelsRK7API.Prop{
		Name:  "Comment",
		Value: Deal.COMMENTS, // TODO без ограничений кол символов
	})
	Props = append(Props, &modelsRK7API.Prop{
		Name:  "Creator",
		Value: Deal.CREATEDBYID, // TODO пока тут только ID, вероятно нужно сделать имя или другую проверку
	})
	Props = append(Props, &modelsRK7API.Prop{
		Name:  "Discount",
		Value: "10", // TODO сделать скидки
	})
	Props = append(Props, &modelsRK7API.Prop{
		Name:  "DealID",
		Value: strconv.Itoa(DealID),
	})

	//заполнить Contact
	if Deal.CONTACTID != "" {
		CONTACTID, err := strconv.Atoi(Deal.CONTACTID)
		if err != nil {
			return errors.Wrapf(err, "failed strconv.Atoi(%s)", Deal.CONTACTID)
		}
		Contact, err := BX24API.ContactGet(CONTACTID)
		if err != nil {
			return errors.Wrapf(err, "failed BX24API.ContactGet(%d)", CONTACTID)
		}
		Props = append(Props, &modelsRK7API.Prop{
			Name:  "Email",
			Value: Contact.EMAIL[0].VALUE, // TODO только всегда первый попавшийся - а что если там будет несколько контактов
		})
		Props = append(Props, &modelsRK7API.Prop{
			Name:  "Phone",
			Value: Contact.PHONE[0].VALUE, // TODO только всегда первый попавшийся - а что если там будет несколько контактов
		})
		Props = append(Props, &modelsRK7API.Prop{
			Name:  "FirstName",
			Value: Contact.NAME,
		})
		Props = append(Props, &modelsRK7API.Prop{
			Name:  "SecondName",
			Value: Contact.SECONDNAME,
		})
		Props = append(Props, &modelsRK7API.Prop{
			Name:  "LastName",
			Value: Contact.LASTNAME,
		})
	}

	Order.ExternalProps.Prop = Props

	//отправить CreateOrder
	resultCreateOrder, err := RK7API.CreateOrder(Order)
	if err != nil {
		logger.Infof("Ошибка при создании заказа RK, error: %v", err)
		return errors.Wrapf(err, "failed RK7API.CreateOrder(%v)", Order)
	}
	logger.Info("Заказ в RK создан успешно")

	//VisitID отправляем в BX24
	err = BX24API.DealUpdate(DealID,
		modelsBX24API.VISITID(fmt.Sprint(resultCreateOrder.VisitID)),
		modelsBX24API.ORDERNAME(resultCreateOrder.Order.OrderName))
	if err != nil {
		logger.Infof("Ошибка при обновлении VisitID=%d, error: %v", resultCreateOrder.VisitID, err)
		return errors.Wrapf(err, "failed BX24API.DealUpdate(DealID: %d, VisitID: %d", DealID, resultCreateOrder.VisitID)
	}
	logger.Infof("VisitID=%d в BX24 обновлен успешно", resultCreateOrder.VisitID)

	//отправить SaveOrder
	//получить ProductRows
	ProductRows, err := BX24API.ProductRowsGet(DealID)
	if err != nil {
		logger.Infof("Ошибка при получении ProductRows, error: %v", err)
		return errors.Wrapf(err, "failed BX24API.ProductRowsGet(%d)", DealID)
	}
	//получить Dishs
	var Dishs []modelsRK7API.Dish
	for _, row := range ProductRows {
		Product, err := BX24API.ProductGet(row.PRODUCTID)
		if err != nil {
			return errors.Wrapf(err, "failed BX24API.ProductGet(%d)", row.PRODUCTID)
		}

		XMLID, err := strconv.Atoi(Product.XMLID)
		if err != nil {
			return errors.Wrapf(err, "failed strconv.Atoi(%s)", Product.XMLID)
		}

		Dishs = append(Dishs, modelsRK7API.Dish{
			ID:       XMLID,
			Quantity: row.QUANTITY * 1000,
		})
	}
	logger.Infof("ProductRows получены, Dishs сформированы, len=%d", len(Dishs))

	//Отправка SaveOrder
	resultSaveOrder, err := RK7API.SaveOrder(resultCreateOrder.VisitID, resultCreateOrder.Guid, cfg.RK7MID.StationCode, &Dishs)
	if err != nil {
		logger.Infof("Ошибка при добавлении блюд в заказ, error: %v", err)
		return errors.Wrapf(err, "failed RK7API.SaveOrder, VisitID: %d, GUID: %s, Station: %d, Dishs: %v", resultCreateOrder.VisitID, resultCreateOrder.Guid, cfg.RK7MID.StationCode, &Dishs)
	}
	logger.Info("В заказ успешно добавлены блюда: %v", Dishs)

	var OrderSum string
	if resultSaveOrder.Order.OrderSum == 0 {
		OrderSum = "0.00"
	} else {
		o := fmt.Sprint(resultSaveOrder.Order.OrderSum)
		OrderSum = fmt.Sprintf("%s.%s", o[:len(o)-2], o[len(o)-2:])
	}

	if OrderSum != Deal.OPPORTUNITY {
		var s []string
		s = append(s, fmt.Sprintf("Сумма заказа RK7=%s не сходится с суммой заказа BX24=%s", OrderSum, Deal.OPPORTUNITY))
		s = append(s, fmt.Sprintf("RK7: VisitID: %d, OrderName: %s, OpenTime: %s", resultSaveOrder.Order.Visit, resultSaveOrder.Order.OrderName, resultSaveOrder.Order.OpenTime))
		s = append(s, fmt.Sprintf("BX24: ID: %d, Name: %s, BEGINDATE:", Deal.ID, Deal.TITLE, Deal.BEGINDATE))
		logger.Infof(strings.Join(s, "\n"))
		return errors.New(strings.Join(s, "\n"))
	}

	logger.Infof("Webhook успешно обработан, Visit: %d, OrderName: %s, OrderSum: %d", resultCreateOrder.VisitID, resultCreateOrder.Order.OrderName, resultSaveOrder.Order.OrderSum)

	//++отправить CreateOrder
	//++отправить SaveOrder

	//TODO сохранить метку в DB, что Webhook закончил работы - чтобы Webhook одновременно не работал с другим webhhok
	//TODO что если будут создаваться много заказов сразу?? - нужна очередь

	//TODO каждые N минут сверять заказ с RK7 и DB
	//TODO СИНХРАААААА!!!!!
	//если заказ RK7 != заказ DB:
	//если удален, то удалить
	//если изменен состав, то изменить состав
	//если закрыт/оплачен, то закрыть

	//если заказы RK7 > заказы DB:
	//то создать заказы в BX24
	//то создать заказы в DB

	//TODO отправить сообщение в telegram что заказ создан успешно

	return nil
}
