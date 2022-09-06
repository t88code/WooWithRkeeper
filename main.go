package main

import (
	"WooWithRkeeper/internal/cache"
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/internal/telegram"
	"WooWithRkeeper/internal/wooapi"
	"WooWithRkeeper/internal/woocommerce"
	"WooWithRkeeper/pkg/logging"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/text/encoding/charmap"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

//запросить меню
//если версия не изменилась, то уснуть на 5 минут
//если версия изменилась, то найти изменения
//найденное изменение отправить в bitrix
//TODO сделать логировование Debug
//TODO сделать архивирование логов
//TODO добавить ID типа цены
//TODO поиск дублей в BX24 по "XML_ID": "1000666"
//TODO поиск дублей в RK7 по ID_BX24
//TODO если папка не активная, то вложенные блюда создаются в корневой папке SectionID = null
//!!!!!TODO добавить проверку при создании заказа что заказ уже есть в DB и что то подобное

func HandlerWebhook(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logger := logging.GetLogger()
	logger.Info("Start HandlerWebhook")
	defer logger.Info("End HandlerWebhook")
	//cfg := config.GetConfig()

	err := r.ParseForm()
	if err != nil {
		err := telegram.SendMessage("Не удалось обработать вебхук на создание заказа")
		if err != nil {
			logger.Errorf("failed telegram.SendMessage(), error: %v", err)
			fmt.Fprint(w, "Error")
			return
		}
		fmt.Fprint(w, "Error")
		return
	}

	logger.Debug("Request\n\t", r)
	logger.Debug("Method\n\t", r.Method)
	logger.Debug("Host\n\t", r.Host)
	logger.Debug("URL\n\t", r.URL)
	logger.Debug("RequestURI\n\t", r.RequestURI)
	logger.Debug("path\n\t", r.URL.Path)
	logger.Debug("Form\n\t", r.Form)
	logger.Debug("MultipartForm\n\t", r.MultipartForm)
	logger.Debug("ContentLength\n\t", r.ContentLength)
	logger.Debug("Header\n\t", r.Header)
	respBody, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		logger.Error(err)
	}
	//Отобразить преобразованный массив байтов в строку

	jsonBody := string(respBody)
	logger.Debug("body\n\t", jsonBody)

	_, err = fmt.Fprint(w, "Ok")
	if err != nil {
		logger.Errorf("failed to send response, error: %v", err)
		return
	}
}

func HandlerWebhookCreateOrder(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logger := logging.GetLogger()
	logger.Info("Start HandlerWebhookCreateOrder")
	defer logger.Info("End HandlerWebhookCreateOrder")
	//cfg := config.GetConfig()

	err := r.ParseForm()
	if err != nil {
		err := telegram.SendMessage("Не удалось обработать вебхук на создание заказа")
		if err != nil {
			logger.Errorf("failed telegram.SendMessage(), error: %v", err)
			fmt.Fprint(w, "Error")
			return
		}
		fmt.Fprint(w, "Error")
		return
	}

	logger.Debug("Request\n\t", r)
	logger.Debug("Method\n\t", r.Method)
	logger.Debug("Host\n\t", r.Host)
	logger.Debug("URL\n\t", r.URL)
	logger.Debug("RequestURI\n\t", r.RequestURI)
	logger.Debug("path\n\t", r.URL.Path)
	logger.Debug("Form\n\t", r.Form)
	logger.Debug("MultipartForm\n\t", r.MultipartForm)
	logger.Debug("ContentLength\n\t", r.ContentLength)
	logger.Debug("Header\n\t", r.Header)
	respBody, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		logger.Error(err)
	}
	//Отобразить преобразованный массив байтов в строку

	logger.Debug("body\n\t", string(respBody))

	err = woocommerce.WebhookCreateOrderInRkeeper(respBody)
	if err != nil {
		errorText := fmt.Sprintf("Не удалось обработать вебхук на создание заказа: %v", err)
		logger.Error(errorText)
		err := telegram.SendMessage(errorText)
		if err != nil {
			logger.Errorf("failed telegram.SendMessage(), error: %v", err)
			fmt.Fprint(w, "Error")
			return
		}
		fmt.Fprint(w, "Error")
		return
	}

	_, err = fmt.Fprint(w, "Ok")
	if err != nil {
		logger.Errorf("failed to send response, error: %v", err)
		return
	}
}

func HandlerPropGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logger := logging.GetLogger()
	logger.Info("Start HandlerPropGet")
	defer logger.Info("End HandlerPropGet")

	cacheOrder := cache.GetCacheOrder()

	visitID, err := strconv.Atoi(ps.ByName("visitid"))
	if err != nil {
		errorText := fmt.Sprintf("не удалось определить visitid=%s: %v", ps.ByName("visitid"), err)
		logger.Error(errorText)
		fmt.Fprintf(w, "")
		return
	}

	logger.Debugf("visitid: %d", visitID)
	logger.Debugf("propname: %s", ps.ByName("propname"))

	order, err := cacheOrder.Get(visitID)
	if err != nil {
		errorText := fmt.Sprintf("не удалось получить из кэша visitid=%d: %v", visitID, err)
		logger.Error(errorText)
		fmt.Fprintf(w, "")
		return
	}

	if order == nil {
		logger.Infof("visitID=%d не найден", visitID)
		fmt.Fprintf(w, "")
		return
	}

	enc := charmap.Windows1251.NewEncoder()

	var propvalue string
	for _, prop := range order.ExternalProps.Prop {
		if prop.Name == ps.ByName("propname") {
			logger.Debugf("propvalue(UTF-8): %s", prop.Value)
			propvalue, err = enc.String(prop.Value)
			if err != nil {
				errorText := fmt.Sprintf("не удалось преобразовать в windows-1251 prop=%s: %v", propvalue, err)
				logger.Error(errorText)
			}
			break
		}
	}

	logger.Debugf("propvalue(win1251): %s", propvalue)

	fmt.Fprintf(w, propvalue)
}

//todo найситься обрабатывать паники

func main() {
	logger := logging.GetLogger()
	logger.Info("Start Main")
	defer logger.Info("End Main")
	var err error

	cfg := config.GetConfig()

	WOOAPI := wooapi.NewAPI(cfg.WOOCOMMERCE.URL, cfg.WOOCOMMERCE.Key, cfg.WOOCOMMERCE.Secret)

	//var p models.Product
	//p.Id = 3124
	//p.Categories = append(p.Categories, &models.Categories{Id: 299})
	//
	//productUpdate, err := WOOAPI.ProductUpdate(&p)
	//if err != nil {
	//	logger.Error(err)
	//	return
	//}
	//
	//logger.Info(productUpdate.Name)
	//logger.Info(productUpdate.Id)
	//logger.Info(productUpdate.Categories)

	//var c models.ProductCategory
	//c.Name = "групаа тестовая 3"
	//c.Parent = cfg.WOOCOMMERCE.MenuCategoryId
	//
	//productCategory, err := WOOAPI.ProductCategoryAdd(&c)
	//if err != nil {
	//	logger.Error(err)
	//	return
	//}
	//logger.Info(productCategory.Name)
	//logger.Info(productCategory.Id)
	//logger.Info(productCategory.Parent)

	//var cNew models.ProductCategory
	//cNew.Id = 300
	//cNew.Parent = cfg.WOOCOMMERCE.MenuCategoryId
	//
	//productCategoryUpdate, err := WOOAPI.ProductCategoryUpdate(&cNew)
	//if err != nil {
	//	logger.Error(err)
	//	return
	//}
	//
	//logger.Info(productCategoryUpdate.Name)
	//logger.Info(productCategoryUpdate.Id)
	//logger.Info(productCategoryUpdate.Parent)
	//
	//panic(2323)
	//
	//_, err = cache.NewCacheMenu()
	//if err != nil {
	//	logger.Errorf("не удалось получить меню: %v", err)
	//	return
	//}

	//RK7API := rk7api.NewAPI(cfg.RK7MID.URL, cfg.RK7.User, cfg.RK7.Pass)
	//getOrderList, err := RK7API.GetOrderList()
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println(123)
	//s := getOrderList.Visit[len(getOrderList.Visit)-1].Orders.Order[0].ExternalID[8].ExtID
	//
	//fmt.Println([]byte(s))
	//
	//panic(234)

	//err = WOOAPI.ProductCategoryDelete(292, options.Force(true))
	//if err != nil {
	//	logger.Error(err)
	//	return
	//}

	//go sync.SyncMenuService()
	go telegram.BotStart()

	//http.HandleFunc("/webhook/creat_order", HandlerWebhookCreateOrder)
	//http.HandleFunc("/", HandlerWebhook)
	//http.HandleFunc("/TransactionsEx", HandlerTransactionsEx)

	//err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.SERVICE.PORT), nil)
	//if err != nil {
	//	errorText := fmt.Sprintf("failed http.ListenAndServe(%d, cfg.SERVICE.PORT), nil), error: %v", cfg.SERVICE.PORT, err)
	//	err := telegram.SendMessage(errorText)
	//	if err != nil {
	//		logger.Errorf("failed telegram.SendMessage(), error: %v", err)
	//	}
	//	logger.Fatal(errorText)
	//}

	router := httprouter.New()
	router.GET("/", HandlerWebhook)
	router.GET("/prop/:visitid/:propname", HandlerPropGet) //todo добавить в скрипте на кассе параметр URL
	router.POST("/webhook/creat_order", HandlerWebhookCreateOrder)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.SERVICE.PORT), router))
}
