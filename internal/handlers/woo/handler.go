package woo

import (
	"WooWithRkeeper/internal/cache"
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/internal/rk7api"
	modelsRK7API "WooWithRkeeper/internal/rk7api/models"
	"WooWithRkeeper/pkg/logging"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
)

type WebhookCreatOrderOld struct {
	Version        string        `json:"version"`
	ShippingTax    string        `json:"shipping_tax"`
	Status         string        `json:"status"`
	CouponLines    []interface{} `json:"coupon_lines"`
	DateCreatedGmt string        `json:"date_created_gmt"`
	Currency       string        `json:"currency"`
	IsEditable     bool          `json:"is_editable"`
	CartTax        string        `json:"cart_tax"`
	DatePaidGmt    interface{}   `json:"date_paid_gmt"`
	MetaData       []struct {
		Key   string `json:"key"`
		Id    int    `json:"id"`
		Value string `json:"value"`
	} `json:"meta_data"`
	DiscountTotal     string        `json:"discount_total"`
	TransactionId     string        `json:"transaction_id"`
	Refunds           []interface{} `json:"refunds"`
	DateCompletedGmt  interface{}   `json:"date_completed_gmt"`
	NeedsProcessing   bool          `json:"needs_processing"`
	PaymentMethod     string        `json:"payment_method"`
	CustomerNote      string        `json:"customer_note"`
	CustomerId        int           `json:"customer_id"`
	ShippingTotal     string        `json:"shipping_total"`
	CustomerUserAgent string        `json:"customer_user_agent"`
	TotalTax          string        `json:"total_tax"`
	CurrencySymbol    string        `json:"currency_symbol"`
	OrderKey          string        `json:"order_key"`
	Id                int           `json:"id"` // Order.ExtID
	DateCompleted     interface{}   `json:"date_completed"`
	ParentId          int           `json:"parent_id"`
	Total             string        `json:"total"`
	Shipping          struct {
		State     string `json:"state"`
		City      string `json:"city"`
		Company   string `json:"company"`
		Phone     string `json:"phone"`
		FirstName string `json:"first_name"`
		Address2  string `json:"address_2"`
		Address1  string `json:"address_1"`
		LastName  string `json:"last_name"`
		Country   string `json:"country"`
		Postcode  string `json:"postcode"`
	} `json:"shipping"`
	NeedsPayment bool `json:"needs_payment"`
	LineItems    []struct {
		Subtotal    string        `json:"subtotal"`
		Taxes       []interface{} `json:"taxes"`
		Quantity    int           `json:"quantity"`
		SubtotalTax string        `json:"subtotal_tax"`
		ParentName  interface{}   `json:"parent_name"`
		Image       struct {
			Id  string `json:"id"`
			Src string `json:"src"`
		} `json:"image"`
		Price    int    `json:"price"`
		Name     string `json:"name"`
		MetaData []struct {
			Key        string `json:"key"`
			DisplayKey string `json:"display_key"`
			Id         int    `json:"id"`
			Value      struct {
				Summary        string `json:"summary,omitempty"`
				Duration       string `json:"duration,omitempty"`
				Persons        string `json:"persons,omitempty"`
				PersonType     string `json:"persontype,omitempty"`     // todo Тип заказчика (Юр.лицо/Физ.лицо)
				CompanyDetails string `json:"companydetails,omitempty"` // todo Реквизиты Юр.лица (если выбрано Юр.лицо)
				Comment        string

				Start struct {
					Timezone     string `json:"timezone"`
					Date         string `json:"date"`
					TimezoneType int    `json:"timezone_type"`
				} `json:"start,omitempty"`
				TimeStart string `json:"timeStart,omitempty"`
				Resources []struct {
					Name            string      `json:"name"`
					Hidden          bool        `json:"hidden"`
					TermId          int         `json:"term_id"`
					Quantity        bool        `json:"quantity"`
					Price           interface{} `json:"price"`
					LimitedResource bool        `json:"limitedResource"`
				} `json:"resources,omitempty"`
				End struct {
					Timezone     string `json:"timezone"`
					Date         string `json:"date"`
					TimezoneType int    `json:"timezone_type"`
				} `json:"end,omitempty"`
				Total int `json:"total,omitempty"`
				Items []struct {
					Label string `json:"label"`
					Price int    `json:"price"`
				} `json:"items,omitempty"`
			} `json:"value"`
			DisplayValue struct {
				Summary  string `json:"summary,omitempty"`
				Duration string `json:"duration,omitempty"`
				Persons  string `json:"persons,omitempty"`
				Start    struct {
					Timezone     string `json:"timezone"`
					Date         string `json:"date"`
					TimezoneType int    `json:"timezone_type"`
				} `json:"start,omitempty"`
				TimeStart string `json:"timeStart,omitempty"`
				Resources []struct {
					Name            string      `json:"name"`
					Hidden          bool        `json:"hidden"`
					TermId          int         `json:"term_id"`
					Quantity        bool        `json:"quantity"`
					Price           interface{} `json:"price"`
					LimitedResource bool        `json:"limitedResource"`
				} `json:"resources,omitempty"`
				End struct {
					Timezone     string `json:"timezone"`
					Date         string `json:"date"`
					TimezoneType int    `json:"timezone_type"`
				} `json:"end,omitempty"`
				Total int `json:"total,omitempty"`
				Items []struct {
					Label string `json:"label"`
					Price int    `json:"price"`
				} `json:"items,omitempty"`
			} `json:"display_value"`
		} `json:"meta_data"`
		TotalTax    string `json:"total_tax"`
		Id          int    `json:"id"`
		ProductId   int    `json:"product_id"`
		TaxClass    string `json:"tax_class"`
		Sku         string `json:"sku"`
		Total       string `json:"total"`
		VariationId int    `json:"variation_id"`
	} `json:"line_items"`
	CustomerIpAddress string        `json:"customer_ip_address"`
	FeeLines          []interface{} `json:"fee_lines"`
	PaymentUrl        string        `json:"payment_url"`
	ShippingLines     []interface{} `json:"shipping_lines"` //+-
	PricesIncludeTax  bool          `json:"prices_include_tax"`
	DateModifiedGmt   string        `json:"date_modified_gmt"`
	DiscountTax       string        `json:"discount_tax"`
	CartHash          string        `json:"cart_hash"`
	Number            string        `json:"number"`
	Links             struct {      //+-
		Self []struct {
			Href string `json:"href"`
		} `json:"self"`
		Collection []struct {
			Href string `json:"href"`
		} `json:"collection"`
	} `json:"_links"`
	DateModified       string        `json:"date_modified"`
	DatePaid           interface{}   `json:"date_paid"`
	TaxLines           []interface{} `json:"tax_lines"`
	PaymentMethodTitle string        `json:"payment_method_title"`
	CreatedVia         string        `json:"created_via"`
	DateCreated        string        `json:"date_created"`
	Billing            struct {
		State     string `json:"state"`
		City      string `json:"city"`
		Company   string `json:"company"`
		Phone     string `json:"phone"`
		FirstName string `json:"first_name"`
		Address2  string `json:"address_2"`
		Address1  string `json:"address_1"`
		LastName  string `json:"last_name"`
		Country   string `json:"country"`
		Email     string `json:"email"`
		Postcode  string `json:"postcode"`
	} `json:"billing"`
}

func WebhookCreateOrderInRKeeper(jsonByteArray []byte) error {

	logger := logging.GetLogger()
	logger.Println("Start WebhookCreateOrderInRKeeper")
	defer logger.Println("End WebhookCreateOrderInRKeeper")
	cfg := config.GetConfig()
	RK7API, err := rk7api.NewAPI(cfg.RK7MID.URL, cfg.RK7.User, cfg.RK7.Pass)
	if err != nil {
		return errors.New("failed rk7api.NewAPI()")
	}

	logger.Info("Запущена обработка Webhook на событие создание заказа")

	WebhookCreatOrder := new(WebhookCreatOrder)
	err = json.Unmarshal(jsonByteArray, WebhookCreatOrder)
	if err != nil {
		return errors.Wrap(err, "failed json.Unmarshal(jsonByteArray, WebhookCreatOrder)")
	}

	//отправить CreateOrder
	//заполнить Order
	Order := new(modelsRK7API.OrderInRK7QueryCreateOrder)
	Order.OrderType = new(modelsRK7API.OrderType)
	Order.OrderType.Code = cfg.RK7MID.OrderTypeCode
	Order.Table = new(modelsRK7API.Table)
	Order.Table.Code = cfg.RK7MID.TableCode
	Order.PersistentComment = "комментарий которыой сохраняемый" //TODO
	Order.ExtSource = "Woocommerce"
	Order.ExtID = strconv.Itoa(WebhookCreatOrder.Id)

	//создать Props в RK7
	//заполнить Props через CreateOrder
	//отобразить Props на кассе
	var Notation []string

	var ID int = WebhookCreatOrder.Id
	var HallName string                                              // Props - HallName - Наименование зала
	var DateStart string                                             // Props - DateStart - Дата бронирования
	var TimeStart string                                             // Props - TimeStart - Время начала пользования залом
	var TimeEnd string                                               // Props - TimeEnd - Время окончания пользования залом
	var Persons string                                               // Props - Persons - Кол-во гостей
	var PersonType string                                            // Props - PersonType - Тип заказчика (Юр.лицо/Физ.лицо)
	var PersonName = fmt.Sprint(WebhookCreatOrder.Billing.FirstName) // Props - PersonName - Имя заказчика
	var LastName string = fmt.Sprint(WebhookCreatOrder.Billing.LastName)
	var CompanyName string = fmt.Sprint(WebhookCreatOrder.Billing.Company) // Props - CompanyName - Наименование Юр.лица (если выбрано Юр.лицо)
	var CompanyDetails string                                              // Props - CompanyDetails - Реквизиты Юр.лица (если выбрано Юр.лицо)
	var Phone string = WebhookCreatOrder.Billing.Phone                     // Props - Phone - Телефон заказчика
	var Email string = WebhookCreatOrder.Billing.Email                     // Props - Email - e-mail заказчика
	var Comment string                                                     // Props - Comment - Комментарий
	var OrderDetails string                                                // Props - OrderDetails - Дополнительные параметры к заказу
	var OrderSum string = WebhookCreatOrder.Total                          // Props - OrderSum - Итоговая стоимость заказа
	var DateCreated string = WebhookCreatOrder.DateCreated                 // Props - DateCreated - Дата оформления заказа
	var Duration string                                                    // Props - Duration - Продолжительность брони
	var DateTimeStart string                                               // OpenTime
	var DurationRK string                                                  //duration="1899-12-30T04:00:00"
	var Deposit int                                                        // "10000"

	var servicesNotation, dishsNotation []string
	//Банкетное меню lite (25000 ₽), Каскад из шампанского (10000 ₽), Ковровая дорожка (3000 ₽)
	//Booking #3049 Unpaid
	//28.07.2022
	//07:00
	//Количество участников: 10
	//Услуги
	//
	//Банкетное меню lite (25000 ₽)
	//Каскад из шампанского (10000 ₽)
	//Ковровая дорожка (3000 ₽)

	//заполнить Props
	var Props []*modelsRK7API.Prop

	Props = append(Props, &modelsRK7API.Prop{
		Name:  "ID",
		Value: strconv.Itoa(ID),
	})
	Props = append(Props, &modelsRK7API.Prop{
		Name:  "Email",
		Value: Email,
	})
	Props = append(Props, &modelsRK7API.Prop{
		Name:  "Phone",
		Value: Phone,
	})
	Props = append(Props, &modelsRK7API.Prop{
		Name:  "PersonName",
		Value: PersonName,
	})
	Props = append(Props, &modelsRK7API.Prop{
		Name:  "LastName",
		Value: LastName,
	})
	Props = append(Props, &modelsRK7API.Prop{
		Name:  "OrderSum",
		Value: OrderSum,
	})
	Props = append(Props, &modelsRK7API.Prop{
		Name:  "CompanyName",
		Value: CompanyName,
	})
	Props = append(Props, &modelsRK7API.Prop{
		Name:  "DateCreated",
		Value: DateCreated,
	})

	//meta_data и приведение
	for _, metaData := range WebhookCreatOrder.MetaData {
		//предоплата
		if metaData.Key == "_wc_deposits_deposit_amount" && metaData.Value != nil {
			if value, ok := metaData.Value.(string); ok {
				Deposit, err = strconv.Atoi(value)
				if err != nil {
					return errors.Wrapf(err, "не удалось конвертировать значение депозита=%s в число", value)
				}
			}
		}
	}

	//line_items
	if len(WebhookCreatOrder.LineItems) > 0 {

		HallName = WebhookCreatOrder.LineItems[0].Name // Props - HallName - Наименование зала
		Props = append(Props, &modelsRK7API.Prop{
			Name:  "HallName",
			Value: HallName,
		})

		if len(WebhookCreatOrder.LineItems[0].MetaData) >= 2 {

			fmt.Println("WebhookCreatOrder.LineItems[0].MetaData", WebhookCreatOrder.LineItems[0].MetaData)

			dateTimeStart, err := time.Parse("2006-01-02 15:04:05.000000", WebhookCreatOrder.LineItems[0].MetaData[0].Value.Start.Date)
			if err != nil {
				return errors.Wrapf(err, "Не удалось распарсить время в поле WebhookCreatOrder.LineItems[0].MetaData[0].Value.Start.Date=%s", WebhookCreatOrder.LineItems[0].MetaData[0].Value.Start.Date)
			}
			dateTimeEnd, err := time.Parse("2006-01-02 15:04:05.000000", WebhookCreatOrder.LineItems[0].MetaData[0].Value.End.Date)
			if err != nil {
				return errors.Wrapf(err, "Не удалось распарсить время в поле WebhookCreatOrder.LineItems[0].MetaData[0].Value.End.Date=%s", WebhookCreatOrder.LineItems[0].MetaData[0].Value.End.Date)
			}

			DateStart = dateTimeStart.Format("2006-01-02") // Props - DateStart - Дата бронирования
			DateTimeStart = dateTimeStart.Format("2006-01-02T15:04:05")
			TimeStart = dateTimeStart.Format("15:04:05")                       // Props - TimeStart - Время начала пользования залом
			TimeEnd = dateTimeEnd.Format("15:04:05")                           // Props - TimeEnd - Время окончания пользования залом
			Persons = WebhookCreatOrder.LineItems[0].MetaData[0].Value.Persons // Props - Persons - Кол-во гостей
			//PersonType = WebhookCreatOrder.LineItems[0].MetaData[0].Value.PersonType                     // Props - PersonType - Тип заказчика (Юр.лицо/Физ.лицо)
			PersonType = ""
			//CompanyDetails = fmt.Sprint(WebhookCreatOrder.LineItems[0].MetaData[0].Value.CompanyDetails) // Props - CompanyDetails - Реквизиты Юр.лица (если выбрано Юр.лицо)
			CompanyDetails = ""
			//Comment = WebhookCreatOrder.LineItems[0].MetaData[0].Value.Comment                           // Props - Comment - Комментарий
			Comment = ""

			durationTime := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
			Duration = WebhookCreatOrder.LineItems[0].MetaData[0].Value.Duration
			durationInt, err := strconv.Atoi(Duration)
			if err != nil {
				return errors.Wrapf(err, "Не удалось распарсить продолжительность брони Duration=%s", Duration)
			}

			DurationRK = durationTime.Add(time.Hour * time.Duration(durationInt)).Format("2006-01-02T15:04:05")
			//duration="1899-12-30T04:00:00"

			Props = append(Props, &modelsRK7API.Prop{
				Name:  "Duration",
				Value: Duration, // TODO TEST
			})
			Props = append(Props, &modelsRK7API.Prop{
				Name:  "DateStart",
				Value: DateStart, // TODO TEST
			})
			Props = append(Props, &modelsRK7API.Prop{
				Name:  "TimeStart",
				Value: TimeStart, // TODO TEST
			})
			Props = append(Props, &modelsRK7API.Prop{
				Name:  "TimeEnd",
				Value: TimeEnd, // TODO TEST
			})
			Props = append(Props, &modelsRK7API.Prop{
				Name:  "Persons",
				Value: Persons, // TODO TEST
			})
			Props = append(Props, &modelsRK7API.Prop{
				Name:  "PersonType",
				Value: PersonType, // TODO TEST
			})
			Props = append(Props, &modelsRK7API.Prop{
				Name:  "CompanyDetails",
				Value: CompanyDetails, // TODO TEST
			})
			Props = append(Props, &modelsRK7API.Prop{
				Name:  "Comment",
				Value: Comment, // TODO TEST
			})

			for _, dish := range WebhookCreatOrder.LineItems[0].MetaData[0].Value.Resources {
				name := fmt.Sprint(dish.Name)
				price := fmt.Sprint(dish.Price)
				dishsNotation = append(dishsNotation, fmt.Sprintf("%s (%s руб)", name, price))
			}

			//TODO сделать проверки, что объект существует
			//TODO сделать проверки, что поле "key":"_mvvwb_order_item_key_costs", "display_key":"_mvvwb_order_item_key_costs",
			for _, item := range WebhookCreatOrder.LineItems[0].MetaData[1].Value.Items {
				label := fmt.Sprint(item.Label)
				price := fmt.Sprint(item.Price)
				servicesNotation = append(servicesNotation, fmt.Sprintf("%s (%s руб)", label, price))
			}
		}
	}

	statusPayed := "Unpaid" //TODO statusPayed

	//собираем примечание
	Notation = append(Notation, fmt.Sprintf("%s", HallName))                       // //  "name":"\u0420\u0435\u0441\u0442\u043e\u0440\u0430\u043d \u00abLe Noir\u00bb", = Ресторан «Le Noir»
	Notation = append(Notation, fmt.Sprintf("Booking №%d, %s", ID, statusPayed))   // TODO > not correct >>> Booking #3049 <<< Unpaid
	Notation = append(Notation, DateStart)                                         // 28.07.2022
	Notation = append(Notation, fmt.Sprintf("%s - %s", TimeStart, TimeEnd))        // 07:00 - 15:00
	Notation = append(Notation, fmt.Sprintf("Количество участников: %s", Persons)) // Количество участников: 10
	Notation = append(Notation, "Услуги:")
	Notation = append(Notation, servicesNotation...)
	//Стоимость бронирования (24000 ₽)
	//Стоимость услуг (38000 ₽)
	Notation = append(Notation, fmt.Sprintf("Итоговая стоимость %s", OrderSum))
	Notation = append(Notation, "Дополнительные услуги:")
	Notation = append(Notation, dishsNotation...)
	//Банкетное меню lite (25000 ₽)
	//Каскад из шампанского (10000 ₽)
	//Ковровая дорожка (3000 ₽)

	OrderDetails = strings.Join(Notation, "\r\n") // Props - OrderDetails - Дополнительные параметры к заказу

	logger.Debugf("OrderDetails:\n%s", OrderDetails)

	Props = append(Props, &modelsRK7API.Prop{
		Name:  "OrderDetails",
		Value: OrderDetails,
	})

	Order.ExternalProps = new(modelsRK7API.ExternalProps)
	Order.ExternalProps.Prop = Props
	Order.Duration = DurationRK
	Order.OpenTime = DateTimeStart

	// TODO Дата начала - сделать проверку что есть
	// TODO Продолжительность - сделать проверку что не 0

	logger.Debug("Order.Table.Code: ", Order.Table.Code)
	logger.Debug("Order.PersistentComment: ", Order.PersistentComment)
	logger.Debug("Order.ExtSource: ", Order.ExtSource)
	logger.Debug("Order.ExtID: ", Order.ExtID)
	logger.Debug("ID: ", ID)
	logger.Debug("HallName: ", HallName)
	logger.Debug("DateStart: ", DateStart)
	logger.Debug("TimeStart: ", TimeStart)
	logger.Debug("TimeEnd: ", TimeEnd)
	logger.Debug("Persons: ", Persons)
	logger.Debug("PersonType: ", PersonType)
	logger.Debug("PersonName: ", PersonName)
	logger.Debug("LastName: ", LastName)
	logger.Debug("CompanyName: ", CompanyName)
	logger.Debug("CompanyDetails: ", CompanyDetails)
	logger.Debug("Phone: ", Phone)
	logger.Debug("Email: ", Email)
	logger.Debug("Comment: ", Comment)
	logger.Debug("OrderDetails: ", OrderDetails)
	logger.Debug("OrderSum: ", OrderSum)
	logger.Debug("DateCreated: ", DateCreated)
	logger.Debug("Duration: ", Duration)
	logger.Debug("DateTimeStart: ", DateTimeStart)
	logger.Debug("DurationRK: ", DurationRK)
	logger.Debug("Deposit: ", Deposit)

	var order *modelsRK7API.Order
	//отправить CreateOrder
	resultCreateOrder, err := RK7API.CreateOrder(Order)
	if err != nil {
		logger.Infof("Ошибка при создании заказа RK, error: %v", err)
		return errors.Wrap(err, "ошибка в RK7API.CreateOrder")
	} else {
		order = resultCreateOrder.Order
		logger.Info("Заказ в RK создан успешно")
	}

	//получить из кэша Order
	cacheOrder := cache.GetCacheOrder() // TODO обновить кеш даже если просто сработа CreateOrder
	err = cacheOrder.Set(order)
	if err != nil {
		return errors.Wrapf(err, "не удалось сохранить заказ (VisitID=%d) в кэше", resultCreateOrder.Order.Visit)
	}
	logger.Info("Заказ успешно сохранен в кэше")

	if Deposit != 0 {
		prepay := new(modelsRK7API.Prepay)
		prepay.Code = cfg.RK7MID.CurrencyCode
		prepay.Amount = Deposit * 100

		logger.Infof("Необходимо добавить предоплату, на сумму %d", Deposit)
		resultSaveOrder, err := RK7API.SaveOrder(resultCreateOrder.Order.Visit,
			resultCreateOrder.Order.Guid,
			cfg.RK7MID.StationCode,
			nil,
			prepay)
		if err != nil {
			return errors.Wrap(err, "ошибка при выполнении SaveOrder")
		} else {
			order = resultSaveOrder.Order
			logger.Info("Предоплата успешно добавлена")
		}
	}

	err = cacheOrder.Set(order)
	if err != nil {
		return errors.Wrapf(err, "не удалось сохранить заказ (VisitID=%d) в кэше", resultCreateOrder.Order.Visit)
	}
	logger.Info("Заказ успешно сохранен в кэше")

	////VisitID отправляем в WOO
	//TODO если не обновим, то не найдем заказ или??
	//допустим будем искать по WOOID -> надо в Props cохранить
	//допустим будем искать по VISITID -> надо в Props cохранить и в WOO>VISITID

	//err = BX24API.DealUpdate(DealID,
	//	modelsBX24API.VISITID(fmt.Sprint(resultCreateOrder.VisitID)),
	//	modelsBX24API.ORDERNAME(resultCreateOrder.Order.OrderName))
	//if err != nil {
	//	logger.Infof("Ошибка при обновлении VisitID=%d, error: %v", resultCreateOrder.VisitID, err)
	//	return errors.Wrapf(err, "failed BX24API.DealUpdate(DealID: %d, VisitID: %d", DealID, resultCreateOrder.VisitID)
	//}
	//logger.Infof("VisitID=%d в BX24 обновлен успешно", resultCreateOrder.VisitID)

	logger.Infof("Webhook успешно обработан, Visit: %d, OrderName: %s", resultCreateOrder.VisitID, resultCreateOrder.Order.OrderName)

	return nil
}
