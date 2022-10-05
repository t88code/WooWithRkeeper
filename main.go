package main

import (
	"WooWithRkeeper/internal/config"
	http2 "WooWithRkeeper/internal/handlers/httphandler"
	"WooWithRkeeper/internal/telegram"
	"WooWithRkeeper/internal/version"
	"WooWithRkeeper/pkg/logging"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
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

//todo найситься обрабатывать паники

func main() {
	logger := logging.GetLogger()
	logger.Info("Start Main")
	v := version.GetVersion()
	logger.Infof("Version %s", v.String())
	defer logger.Info("End Main")
	//var err error

	cfg := config.GetConfig()

	//go sync.SyncMenuService()
	go telegram.BotStart()

	router := httprouter.New()
	router.GET("/", http2.HandlerOtherAll)
	router.POST("/webhook/creat_order", http2.HandlerWebhookCreateOrder)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.SERVICE.PORT), router))
}
