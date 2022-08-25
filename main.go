package main

import (
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/internal/farcards"
	"WooWithRkeeper/internal/sync"
	"WooWithRkeeper/internal/telegram"
	"WooWithRkeeper/internal/webhook"
	"WooWithRkeeper/internal/woocommerce"
	"WooWithRkeeper/pkg/logging"
	"encoding/xml"
	"fmt"
	"io/ioutil"
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

func HandlerRobotsWebhookCreateDeal(w http.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()
	logger.Info("Start HandlerRobotsWebhookCreateDeal")
	defer logger.Info("End HandlerRobotsWebhookCreateDeal")
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

	dealid := r.Form.Get("deal_id")

	logger.Debugf("URL: %s", r.URL)
	logger.Debugf("DealID: %s", dealid)
	logger.Debugf("Form: %v", r.Form)

	DealIDint, err := strconv.Atoi(dealid)
	if err != nil {
		logger.Errorf(fmt.Sprintf("failed strconv.Atoi(%s)", dealid))
		err := telegram.SendMessage(fmt.Sprintf("Не удалось обработать вебхук на создание заказа, ошибка в ID сделки: DealID=%s", dealid))
		if err != nil {
			logger.Errorf("failed telegram.SendMessage(), error: %v", err)
			fmt.Fprint(w, "Error")
			return
		}
		fmt.Fprint(w, "Error")
		return
	}

	err = webhook.CreateDealInRkeeper(DealIDint)
	if err != nil {
		logger.Errorf("failed webhook.CreateDealInRkeeper, error: %v", err)
		err := telegram.SendMessage(fmt.Sprintf("Не удалось обработать вебхук на создание заказа, ошибка: %v", err))
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

func HandlerTransactionsEx(w http.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()
	logger.Info("Start Handler")
	defer logger.Info("End Handler")

	err := r.ParseForm()
	if err != nil {
		err := telegram.SendMessage("Не удалось обработать оплату/удаление на кассе")
		if err != nil {
			logger.Errorf("failed telegram.SendMessage(), error: %v", err)
			fmt.Fprint(w, "Error")
			return
		}
		fmt.Fprint(w, "Error")
		return
	}

	logger.Debugln(r.Form) // print form information in server side
	logger.Debugln("Request\n\t", r)
	logger.Debugln("Method\n\t", r.Method)
	logger.Debugln("Host\n\t", r.Host)
	logger.Debugln("URL\n\t", r.URL)
	logger.Debugln("RequestURI\n\t", r.RequestURI)
	logger.Debugln("path\n\t", r.URL.Path)
	logger.Debugln("Form\n\t", r.Form)
	logger.Debugln("MultipartForm\n\t", r.MultipartForm)
	logger.Debugln("ContentLength\n\t", r.ContentLength)
	logger.Debugln("Header\n\t", r.Header)

	respBody, err := ioutil.ReadAll(r.Body)
	err = r.Body.Close()
	if err != nil {
		logger.Errorf("failed r.Body.Close(), error: %v", err)
	}

	logger.Debugf("body:\n%s", string(respBody))

	Transaction := new(farcards.Transaction)
	err = xml.Unmarshal(respBody, Transaction)
	if err != nil {
		logger.Errorf("failed xml.Unmarshal(respBody, Transaction), error: %v", err)
	}

	logger.Infof("Получен заказ, Guid: %s, OrderName: %s, CheckNum: %d, Sum: %d", Transaction.CHECKDATA.Orderguid, Transaction.CHECKDATA.Ordernum, Transaction.CHECKDATA.Checknum, Transaction.CHECKDATA.CHECKCATEGS.CATEG.Sum)
	err = sync.HandlerTransaction(Transaction)
	if err != nil {
		logger.Error("failed sync.HandlerTransaction(Transaction), error: %v", err)
		err := telegram.SendMessage(fmt.Sprintf("Не удалось обработать транзакцию с кассы, error: %v", err))
		if err != nil {
			logger.Errorf("failed telegram.SendMessage(), error: %v", err)
			fmt.Fprint(w, "Error")
			return
		}
	}

	_, err = fmt.Fprint(w, "Ok")
	if err != nil {
		logger.Errorf("failed to send response, error: %v", err)
		return
	}
}

func HandlerWebhook(w http.ResponseWriter, r *http.Request) {
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

func HandlerWebhookCreateOrder(w http.ResponseWriter, r *http.Request) {
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
		err := telegram.SendMessage("Не удалось обработать вебхук на создание заказа")
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

func main() {
	logger := logging.GetLogger()
	logger.Info("Start Main")
	defer logger.Info("End Main")

	cfg := config.GetConfig()

	//go sync.SyncMenuService()
	go telegram.BotStart()

	http.HandleFunc("/webhook/creat_order", HandlerWebhookCreateOrder)
	http.HandleFunc("/", HandlerWebhook)
	//http.HandleFunc("/TransactionsEx", HandlerTransactionsEx)

	err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.SERVICE.PORT), nil)
	if err != nil {
		errorText := fmt.Sprintf("failed http.ListenAndServe(%d, cfg.SERVICE.PORT), nil), error: %v", cfg.SERVICE.PORT, err)
		err := telegram.SendMessage(errorText)
		if err != nil {
			logger.Errorf("failed telegram.SendMessage(), error: %v", err)
		}
		logger.Fatal(errorText)
	}

}
