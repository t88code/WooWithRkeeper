package sync

import (
	"WooWithRkeeper/internal/bx24api"
	modelsBX24API "WooWithRkeeper/internal/bx24api/models"
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/internal/farcards"
	"WooWithRkeeper/internal/rk7api"
	modelsRK7API "WooWithRkeeper/internal/rk7api/models"
	"WooWithRkeeper/pkg/logging"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"strconv"
)

const (
	STAGE_ID_WON  = "WON"
	STAGE_ID_LOSE = "LOSE"
)

// обработка транзакций при Оплате/Удалении заказа RK
func HandlerTransaction(tr *farcards.Transaction) error {

	logger := logging.GetLogger()
	logger.Println("Start HandlerTransaction")
	defer logger.Println("End HandlerTransaction")

	cfg := config.GetConfig()
	RK7API := rk7api.NewAPI(cfg.RK7MID.URL, cfg.RK7MID.User, cfg.RK7MID.Pass)
	BX24API := bx24api.NewAPI(cfg.BX24.URL)
	db, err := sqlx.Connect("sqlite3", cfg.DBSQLITE.DB)

	if err != nil {
		return errors.Wrap(err, "failed sqlx.Connect")
	}
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			logger.Infof("failed close sqlx.Connect, error: %v", err)
		}
	}(db)

	logger.Infof("Запрашиваем инфо о заказе из RK7")
	rk7QueryResultGetOrder, err := RK7API.GetOrder(tr.CHECKDATA.Orderguid)
	if err != nil {
		return errors.Wrapf(err, "failed RK7API.GetOrder(%s)", tr.CHECKDATA.Orderguid)
	}

	visit := rk7QueryResultGetOrder.Order.Visit
	logger.Infof("Заказ найден в RK7, visit = %d", visit)
	logger.Info("Выполняем поиск сделки в BX24 по visit")
	deals, err := BX24API.DealList(modelsBX24API.Filter(cfg.BX24.FieldVISITID, fmt.Sprint(visit)))
	if err != nil {
		return errors.Wrapf(err, "failed BX24API.DealList(%s)", visit)
	}

	var STAGEID string
	switch tr.Chmode {
	case 1: //заказ оплачен
		STAGEID = STAGE_ID_WON
	case 3: //заказ удален
		STAGEID = STAGE_ID_LOSE
	default:
		return errors.New(fmt.Sprintf("Не удалось определить у транзакции Chmode=%d", tr.Chmode))
	}

	//сделка не найдена в BX24

	switch len(deals) {
	case 0:
		logger.Info("Сделка не найдена в BX24. Приступаем к созданию сделки.")
		//создаем сделку в BX24
		dealID, err := BX24API.DealAdd(
			modelsBX24API.TITLE(fmt.Sprintf("Заказ %s", tr.CHECKDATA.Ordernum)),
			modelsBX24API.VISITID(fmt.Sprint(visit)),
			modelsBX24API.ORDERNAME(rk7QueryResultGetOrder.Order.OrderName),
			modelsBX24API.STAGEID(STAGEID),
		)
		if err != nil {
			return errors.Wrapf(err, "failed BX24API.DealAdd()")
		}
		logger.Infof("Сделка успешно создана, ID: %d", dealID)

		err = UpdateProductRowsInDeal(dealID, rk7QueryResultGetOrder.Order, BX24API)
		if err != nil {
			return errors.Wrapf(err, "failed UpdateProductRowsInDeal(%d)", dealID)
		}

		//TODO email phone и всякую хрень добавить - после закрытия визита ничего не добавить, поэтому UpdateOrder - отменяем
		//возможно можно сделать перед оплатой скриптом на форме
		//<?xml version="1.0" encoding="utf-8"?>
		//<RK7QueryResult ServerVersion="7.6.4.483" XmlVersion="248" NetName="TEST" Status="Query Executing Error" CMD="UpdateOrder" ErrorText="Визит 1159 закрыт!" DateTime="2022-08-21T17:59:19" WorkTime="0" RK7ErrorN="2128" Processed="0" ArrivalDateTime="2022-08-21T17:59:19"/>
		//guid := rk7QueryResultGetOrder.Order.Guid
		//_, err = RK7API.UpdateOrder(guid, modelsRK7API.ExternalProp("DealID", strconv.Itoa(dealID)))
		//if err != nil {
		//	return errors.Wrapf(err, "failed RK7API.UpdateOrder(GUID: %s)", guid)
		//}

		//Обновить DB таблицу Orders
	case 1:
		//если сделка найдена, то обновляем всю инфу
		logger.Infof("Сделка найдена в BX24, DealID: %s. Приступаем к сверке сделки.", deals[0].ID)
		//обновляем сделку в BX24
		dealID, err := strconv.Atoi(deals[0].ID)
		if err != nil {
			return errors.Wrapf(err, "failed strconv.Atoi(DealID: %s)", deals[0].ID)
		}

		//обновляем OrderName, Status
		err = BX24API.DealUpdate(dealID,
			modelsBX24API.ORDERNAME(rk7QueryResultGetOrder.Order.OrderName),
			modelsBX24API.STAGEID(STAGEID),
		)
		if err != nil {
			return errors.Wrapf(err, "failed BX24API.DealUpdate(dealID: %d, OrderName: %s, StageID: %s", dealID, rk7QueryResultGetOrder.Order.OrderName, STAGEID)
		}

		//проверить состав
		//получить состав из сделки BX24
		productRows, err := BX24API.ProductRowsGet(dealID)
		if err != nil {
			return errors.Wrapf(err, "failed BX24API.ProductRowsGet(dealID: %d)", dealID)
		}

		//формируем состав Dishs из заказа RK
		var dishs []*modelsRK7API.Dish
		for _, session := range rk7QueryResultGetOrder.Order.Session {
			for i, _ := range session.Dish {
				dishs = append(dishs, session.Dish[i])
			}
		}
		//выполняем сверку ProductRows<>Orsers.Session.Dishs
		logger.Infof("Выполняем сверку состава заказа")
		if len(productRows) != len(dishs) {
			logger.Infof("Состав блюд различается по длине")
			err = UpdateProductRowsInDeal(dealID, rk7QueryResultGetOrder.Order, BX24API)
			if err != nil {
				return errors.Wrapf(err, "failed UpdateProductRowsInDeal(%d)", dealID)
			}
		} else {
			logger.Infof("Состав блюд совпадает по длине")
			logger.Infof("Выполняем поблюдную сверку")

			//получить актуальное меню RK7
			logger.Println("Получить список всех блюд из RK7")
			Rk7QueryResultGetRefDataMenuitems, err := RK7API.GetRefData("Menuitems",
				modelsRK7API.OnlyActive("0"), //неактивные будем менять статус на N ?или может удалять? в bitrix24
				modelsRK7API.IgnoreEnums("1"),
				modelsRK7API.WithChildItems("3"),
				modelsRK7API.WithMacroProp("1"),
				modelsRK7API.PropMask("items.(Code,Name,Ident,ItemIdent,GUIDString,MainParentIdent,ExtCode,PRICETYPES^3,CategPath,Status,genIDBX24,genSectionIDBX24)"))
			if err != nil {
				return errors.Wrap(err, "Ошибка при выполнении rk7api.GetRefData")
			}
			MenuInRK7 := (Rk7QueryResultGetRefDataMenuitems).(*modelsRK7API.RK7QueryResultGetRefDataMenuitems)
			logger.Printf("Длина списка MenuInRK7 = %d\n", len(MenuInRK7.RK7Reference.Items.Item))
			//сформировать MenuInRK7Map
			MenuInRK7Map := make(map[int]modelsRK7API.MenuitemItem)
			for i, item := range MenuInRK7.RK7Reference.Items.Item {
				MenuInRK7Map[item.ItemIdent] = MenuInRK7.RK7Reference.Items.Item[i]
			}
			logger.Printf("Длина списка MenuInRK7Map = %d\n", len(MenuInRK7Map))

			for i, _ := range dishs {
				if productRows[i].PRODUCTID != MenuInRK7Map[dishs[i].ID].ID_BX24 ||
					productRows[i].PRICE != dishs[i].Price/100 ||
					productRows[i].QUANTITY != dishs[i].Quantity/1000 {

					logger.Infof("Состава заказа различается между BX24 и RK. Запускаем процесс обновления состава сделки.")
					err = UpdateProductRowsInDeal(dealID, rk7QueryResultGetOrder.Order, BX24API)
					if err != nil {
						return errors.Wrapf(err, "failed UpdateProductRowsInDeal(%d)", dealID)
					}
					break
				}
			}
		}
		logger.Infof("Сверка состава заказа выполнена успешно")
	default:
		//если найдено несколько сделок, то что то пошло не так и отправить сообщение в телеграм
		logger.Info("Найдено более одной сделки с visit: %d. Неизвестное состояние", visit)
		return errors.New("в системе BX24 найдено несколько сделок")
	}
	return nil
}

//обновление сделки на основании заказа
func UpdateProductRowsInDeal(dealID int, Order *modelsRK7API.Order, BX24API bx24api.BX24API) error {
	logger := logging.GetLogger()
	logger.Println("Start UpdateProductRowsInDeal")
	defer logger.Println("End UpdateProductRowsInDeal")

	logger.Infof("Запущен процесс обновления ProductRows")
	var products []modelsBX24API.Row
	for _, session := range Order.Session {
		for _, dish := range session.Dish {
			logger.Debug("Добавляем блюдо в ProductSection:")

			productList, err := BX24API.ProductList(modelsBX24API.Filter("XML_ID", strconv.Itoa(dish.ID))) //TODO оптимизировать через ProductListMap
			if err != nil {
				return errors.Wrapf(err, "failed BX24API.ProductList(Filter(XML_ID))")
			}
			if len(productList) == 0 {
				return errors.New(fmt.Sprintf("dish not found in Product BX24, RK_ID: %d, PRICE: %d, QUANTITY: %d", dish.ID, dish.Price, dish.Quantity))
			}

			logger.Debugf("Найден Product: Name: %s, ProductID: %s, RK_ID: %s, PRICE: %s",
				productList[0].NAME,
				productList[0].ID,
				productList[0].XMLID,
				productList[0].PRICE)

			//ProductID from BX24
			productid, err := strconv.Atoi(productList[0].ID)
			if err != nil {
				return err
			}

			// RK: 4000 > BX24: 40.00
			p := fmt.Sprint(dish.Price)
			price := fmt.Sprintf("%s.%s", p[:len(p)-2], p[len(p)-2:])

			// RK: 4000 > BX24: 4
			quantity := dish.Quantity / 1000

			products = append(products, modelsBX24API.PRODUCT(productid, price, quantity))
			logger.Debug("Блюдо добавлено успешно в ProductSection:")
		}
	}

	err := BX24API.ProductRowsSet(dealID, products...)
	if err != nil {
		return errors.Wrapf(err, "failed BX24API.ProductRowsSet(%d)", dealID)
	}
	logger.Infof("ProductRows успешно обновлен")
	return nil
}
