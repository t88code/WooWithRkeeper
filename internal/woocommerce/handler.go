package woocommerce

import (
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/internal/rk7api"
	modelsRK7API "WooWithRkeeper/internal/rk7api/models"
	"WooWithRkeeper/pkg/logging"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

type WebhookCreatOrder struct {
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
	ShippingLines     []interface{} `json:"shipping_lines"`
	PricesIncludeTax  bool          `json:"prices_include_tax"`
	DateModifiedGmt   string        `json:"date_modified_gmt"`
	DiscountTax       string        `json:"discount_tax"`
	CartHash          string        `json:"cart_hash"`
	Number            string        `json:"number"`
	Links             struct {
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

const exampleJsonString = `
{
   "version":"6.6.1",
   "shipping_tax":"0",
   "status":"processing",
   "coupon_lines":[
      
   ],
   "date_created_gmt":"2022-07-14T12:38:51",
   "currency":"RUB",
   "is_editable":false,
   "cart_tax":"0",
   "date_paid_gmt":null,
   "meta_data":[
      {
         "key":"is_vat_exempt",
         "id":42558,
         "value":"no"
      },
      {
         "key":"_new_order_email_sent",
         "id":42582,
         "value":"true"
      }
   ],
   "discount_total":"0",
   "transaction_id":"",
   "refunds":[
      
   ],
   "date_completed_gmt":null,
   "needs_processing":true,
   "payment_method":"cod",
   "customer_note":"",
   "customer_id":0,
   "shipping_total":"0",
   "customer_user_agent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36",
   "total_tax":"0",
   "currency_symbol":"\u20bd",
   "order_key":"wc_order_50et7a59RtvMA",
   "id":3048,
   "date_completed":null,
   "parent_id":0,
   "total":"62000",
   "shipping":{
      "state":"",
      "city":"",
      "company":"",
      "phone":"",
      "first_name":"",
      "address_2":"",
      "address_1":"",
      "last_name":"",
      "country":"",
      "postcode":""
   },
   "needs_payment":false,
   "line_items":[
      {
         "subtotal":"62000",
         "taxes":[
            
         ],
         "quantity":1,
         "subtotal_tax":"0",
         "parent_name":null,
         "image":{
            "id":"",
            "src":""
         },
         "price":62000,
         "name":"\u0420\u0435\u0441\u0442\u043e\u0440\u0430\u043d \u00abLe Noir\u00bb",
         "meta_data":[
            {
               "key":"_mvvwb_order_item_key",
               "display_key":"_mvvwb_order_item_key",
               "id":106,
               "value":{
                  "summary":"28.07.2022  07:00 - 15:00",
                  "duration":"8",
                  "persons":"10",
                  "start":{
                     "timezone":"+04:00",
                     "date":"2022-07-28 07:00:00.000000",
                     "timezone_type":1
                  },
                  "timeStart":"420",
                  "resources":[
                     {
                        "name":"\u0411\u0430\u043d\u043a\u0435\u0442\u043d\u043e\u0435 \u043c\u0435\u043d\u044e lite",
                        "hidden":false,
                        "term_id":93,
                        "quantity":false,
                        "price":25000,
                        "limitedResource":false
                     },
                     {
                        "name":"\u041a\u0430\u0441\u043a\u0430\u0434 \u0438\u0437 \u0448\u0430\u043c\u043f\u0430\u043d\u0441\u043a\u043e\u0433\u043e",
                        "hidden":false,
                        "term_id":92,
                        "quantity":false,
                        "price":"10000",
                        "limitedResource":false
                     },
                     {
                        "name":"\u041a\u043e\u0432\u0440\u043e\u0432\u0430\u044f \u0434\u043e\u0440\u043e\u0436\u043a\u0430",
                        "hidden":false,
                        "term_id":91,
                        "quantity":false,
                        "price":"3000",
                        "limitedResource":false
                     }
                  ],
                  "end":{
                     "timezone":"+04:00",
                     "date":"2022-07-28 15:00:00.000000",
                     "timezone_type":1
                  }
               },
               "display_value":{
                  "summary":"28.07.2022  07:00 - 15:00",
                  "duration":"8",
                  "persons":"10",
                  "start":{
                     "timezone":"+04:00",
                     "date":"2022-07-28 07:00:00.000000",
                     "timezone_type":1
                  },
                  "timeStart":"420",
                  "resources":[
                     {
                        "name":"\u0411\u0430\u043d\u043a\u0435\u0442\u043d\u043e\u0435 \u043c\u0435\u043d\u044e lite",
                        "hidden":false,
                        "term_id":93,
                        "quantity":false,
                        "price":25000,
                        "limitedResource":false
                     },
                     {
                        "name":"\u041a\u0430\u0441\u043a\u0430\u0434 \u0438\u0437 \u0448\u0430\u043c\u043f\u0430\u043d\u0441\u043a\u043e\u0433\u043e",
                        "hidden":false,
                        "term_id":92,
                        "quantity":false,
                        "price":"10000",
                        "limitedResource":false
                     },
                     {
                        "name":"\u041a\u043e\u0432\u0440\u043e\u0432\u0430\u044f \u0434\u043e\u0440\u043e\u0436\u043a\u0430",
                        "hidden":false,
                        "term_id":91,
                        "quantity":false,
                        "price":"3000",
                        "limitedResource":false
                     }
                  ],
                  "end":{
                     "timezone":"+04:00",
                     "date":"2022-07-28 15:00:00.000000",
                     "timezone_type":1
                  }
               }
            },
            {
               "key":"_mvvwb_order_item_key_costs",
               "display_key":"_mvvwb_order_item_key_costs",
               "id":107,
               "value":{
                  "total":62000,
                  "items":[
                     {
                        "label":"\u0421\u0442\u043e\u0438\u043c\u043e\u0441\u0442\u044c \u0431\u0440\u043e\u043d\u0438\u0440\u043e\u0432\u0430\u043d\u0438\u044f",
                        "price":24000
                     },
                     {
                        "label":"\u0421\u0442\u043e\u0438\u043c\u043e\u0441\u0442\u044c \u0443\u0441\u043b\u0443\u0433",
                        "price":38000
                     }
                  ]
               },
               "display_value":{
                  "total":62000,
                  "items":[
                     {
                        "label":"\u0421\u0442\u043e\u0438\u043c\u043e\u0441\u0442\u044c \u0431\u0440\u043e\u043d\u0438\u0440\u043e\u0432\u0430\u043d\u0438\u044f",
                        "price":24000
                     },
                     {
                        "label":"\u0421\u0442\u043e\u0438\u043c\u043e\u0441\u0442\u044c \u0443\u0441\u043b\u0443\u0433",
                        "price":38000
                     }
                  ]
               }
            }
         ],
         "total_tax":"0",
         "id":11,
         "product_id":198,
         "tax_class":"",
         "sku":"",
         "total":"62000",
         "variation_id":0
      }
   ],
   "customer_ip_address":"185.76.222.2",
   "fee_lines":[
      
   ],
   "payment_url":"http://new.hotelslovakia.ru/checkout/order-pay/3048/?pay_for_order=true&key=wc_order_50et7a59RtvMA",
   "shipping_lines":[
      
   ],
   "prices_include_tax":false,
   "date_modified_gmt":"2022-07-14T12:38:51",
   "discount_tax":"0",
   "cart_hash":"f7a43d1bd1959362c234249c8700613e",
   "number":"3048",
   "_links":{
      "self":[
         {
            "href":"http://new.hotelslovakia.ru/wp-json/wc/v3/orders/3048"
         }
      ],
      "collection":[
         {
            "href":"http://new.hotelslovakia.ru/wp-json/wc/v3/orders"
         }
      ]
   },
   "date_modified":"2022-07-14T16:38:51",
   "date_paid":null,
   "tax_lines":[
      
   ],
   "payment_method_title":"\u041e\u043f\u043b\u0430\u0442\u0430 \u043f\u0440\u0438 \u0434\u043e\u0441\u0442\u0430\u0432\u043a\u0435",
   "created_via":"checkout",
   "date_created":"2022-07-14T16:38:51",
   "billing":{
      "state":"",
      "city":"",
      "company":"",
      "phone":"89991155811",
      "first_name":"",
      "address_2":"",
      "address_1":"",
      "last_name":"",
      "country":"",
      "email":"elmeevru@yandex.ru",
      "postcode":""
   }
}

`

func WebhookCreateOrderInRkeeper(jsonByteArray []byte) error {

	logger := logging.GetLogger()
	logger.Println("Start WebhookCreateOrderInRkeeper")
	defer logger.Println("End WebhookCreateOrderInRkeeper")
	cfg := config.GetConfig()
	RK7API := rk7api.NewAPI(cfg.RK7MID.URL, cfg.RK7.User, cfg.RK7.Pass)

	logger.Info("Запущена обработка Webhook на событие создание заказа")

	WebhookCreatOrder := new(WebhookCreatOrder)
	err := json.Unmarshal(jsonByteArray, WebhookCreatOrder)
	if err != nil {
		return errors.Wrap(err, "failed json.Unmarshal(jsonByteArray, WebhookCreatOrder)")
	}

	//отправить CreateOrder
	//заполнить Order
	Order := new(modelsRK7API.OrderInRK7QueryCreateOrder)
	Order.OrderType.Code = cfg.RK7MID.OrderTypeCode
	Order.Table.Code = cfg.RK7MID.TableCode
	Order.PersistentComment = "PersistentComment" //TODO
	Order.ExtSource = "Woocommerce"
	Order.ExtID = strconv.Itoa(WebhookCreatOrder.Id)
	Order.Reserv = 1

	//создать Props в RK7
	//заполнить Props через CreateOrder
	//отобразить Props на кассе
	var notation []string
	var title, id, persons, sum, summary, statusPayed, duration string
	var servicesNotation, dishsNotation []string //Банкетное меню lite (25000 ₽), Каскад из шампанского (10000 ₽), Ковровая дорожка (3000 ₽)
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
		Name:  "Email",
		Value: WebhookCreatOrder.Billing.Email,
	})
	Props = append(Props, &modelsRK7API.Prop{
		Name:  "Phone",
		Value: WebhookCreatOrder.Billing.Phone,
	})
	Props = append(Props, &modelsRK7API.Prop{
		Name:  "FirstName",
		Value: fmt.Sprint(WebhookCreatOrder.Billing.FirstName),
	})
	Props = append(Props, &modelsRK7API.Prop{
		Name:  "LastName",
		Value: fmt.Sprint(WebhookCreatOrder.Billing.LastName),
	})
	sum = WebhookCreatOrder.Total // "total":"62000",
	Props = append(Props, &modelsRK7API.Prop{
		Name:  "Sum",
		Value: sum,
	})
	id = strconv.Itoa(WebhookCreatOrder.Id) // "id":3048,
	Props = append(Props, &modelsRK7API.Prop{
		Name:  "ID",
		Value: id,
	})

	//line_items
	if len(WebhookCreatOrder.LineItems) > 0 {
		title = fmt.Sprint(WebhookCreatOrder.LineItems[0].Name) //  "name":"\u0420\u0435\u0441\u0442\u043e\u0440\u0430\u043d \u00abLe Noir\u00bb",
		Props = append(Props, &modelsRK7API.Prop{
			Name:  "Title",
			Value: title,
		})
		if len(WebhookCreatOrder.LineItems[0].MetaData) == 2 {

			summary = WebhookCreatOrder.LineItems[0].MetaData[0].Value.Summary   // "summary":"28.07.2022  07:00 - 15:00",
			persons = WebhookCreatOrder.LineItems[0].MetaData[0].Value.Persons   // "persons":"10",
			duration = WebhookCreatOrder.LineItems[0].MetaData[0].Value.Duration // "duration":"8",

			Props = append(Props, &modelsRK7API.Prop{
				Name:  "Summary",
				Value: summary,
			})
			Props = append(Props, &modelsRK7API.Prop{
				Name:  "Persons",
				Value: persons, // TODO createOrder attr
			})
			Props = append(Props, &modelsRK7API.Prop{
				Name:  "Duration",
				Value: duration, // TODO createOrder attr
			})

			for _, dish := range WebhookCreatOrder.LineItems[0].MetaData[0].Value.Resources {
				name := fmt.Sprint(dish.Name)
				price := fmt.Sprint(dish.Price)
				dishsNotation = append(dishsNotation, fmt.Sprintf("%s (%s ₽)", name, price))
			}

			//TODO сделать проверки, что объект существует
			//TODO сделать проверки, что поле "key":"_mvvwb_order_item_key_costs", "display_key":"_mvvwb_order_item_key_costs",
			for _, item := range WebhookCreatOrder.LineItems[0].MetaData[1].Value.Items {
				label := fmt.Sprint(item.Label)
				price := fmt.Sprint(item.Price)
				servicesNotation = append(servicesNotation, fmt.Sprintf("%s (%s ₽)", label, price))
			}
		}
	}

	statusPayed = "Unpaid" //TODO statusPayed
	//собираем примечание
	notation = append(notation, fmt.Sprintf("%s", title))                        // //  "name":"\u0420\u0435\u0441\u0442\u043e\u0440\u0430\u043d \u00abLe Noir\u00bb", = Ресторан «Le Noir»
	notation = append(notation, fmt.Sprintf("Booking #%s, %s", id, statusPayed)) // TODO > not correct >>> Booking #3049 <<< Unpaid
	//notation = append(notation, "\n")
	notation = append(notation, summary)                                           // 28.07.2022  07:00 - 15:00
	notation = append(notation, fmt.Sprintf("Количество участников: %s", persons)) // Количество участников: 10
	//notation = append(notation, "\n")
	notation = append(notation, "Услуги:")
	notation = append(notation, servicesNotation...)
	//Стоимость бронирования (24000 ₽)
	//Стоимость услуг (38000 ₽)
	notation = append(notation, fmt.Sprintf("Итоговая стоимость %s", sum))
	notation = append(notation, "Дополнительные услуги:")
	notation = append(notation, dishsNotation...)
	//Банкетное меню lite (25000 ₽)
	//Каскад из шампанского (10000 ₽)
	//Ковровая дорожка (3000 ₽)

	logger.Debugf("Notation:\n%s", strings.Join(notation, "\n"))

	Props = append(Props, &modelsRK7API.Prop{
		Name:  "Notation",
		Value: strings.Join(notation, "\n"),
	})
	Order.ExternalProps.Prop = Props

	//отправить CreateOrder
	resultCreateOrder, err := RK7API.CreateOrder(Order)
	if err != nil {
		logger.Infof("Ошибка при создании заказа RK, error: %v", err)
		return errors.Wrapf(err, "failed RK7API.CreateOrder(%v)", Order)
	}

	logger.Info("Заказ в RK создан успешно")

	////VisitID отправляем в BX24
	//err = BX24API.DealUpdate(DealID,
	//	modelsBX24API.VISITID(fmt.Sprint(resultCreateOrder.VisitID)),
	//	modelsBX24API.ORDERNAME(resultCreateOrder.Order.OrderName))
	//if err != nil {
	//	logger.Infof("Ошибка при обновлении VisitID=%d, error: %v", resultCreateOrder.VisitID, err)
	//	return errors.Wrapf(err, "failed BX24API.DealUpdate(DealID: %d, VisitID: %d", DealID, resultCreateOrder.VisitID)
	//}
	//logger.Infof("VisitID=%d в BX24 обновлен успешно", resultCreateOrder.VisitID)

	logger.Infof("Webhook успешно обработан, Visit: %d, OrderName: %s", resultCreateOrder.VisitID, resultCreateOrder.Order.OrderName)

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
