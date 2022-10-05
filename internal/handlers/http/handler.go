package http

import (
	"WooWithRkeeper/internal/handlers/woo"
	"WooWithRkeeper/internal/telegram"
	"WooWithRkeeper/internal/version"
	"WooWithRkeeper/pkg/logging"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"net/http"
)

func HandlerOtherAll(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	logger := logging.GetLogger()
	logger.Info("Start HandlerOtherAll")
	defer logger.Info("End HandlerOtherAll")
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

	v := version.GetVersion()
	_, err = fmt.Fprintf(w, "Version %s", v.String())
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

	err = woo.WebhookCreateOrderInRkeeper(respBody)
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
