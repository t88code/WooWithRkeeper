package main

import (
	"WooWithRkeeper/internal/cache"
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/internal/database"
	"WooWithRkeeper/internal/handlers/httphandler"
	"WooWithRkeeper/internal/license"
	"WooWithRkeeper/internal/rk7api"
	"WooWithRkeeper/internal/sync"
	"WooWithRkeeper/internal/telegram"
	"WooWithRkeeper/internal/version"
	"WooWithRkeeper/internal/wooapi"
	"WooWithRkeeper/pkg/logging"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

//TODO сделать логировование Debug
//TODO сделать архивирование логов
//TODO добавить ID типа цены
//TODO поиск дублей в BX24 по "XML_ID": "1000666"
//TODO поиск дублей в RK7 по ID_BX24
//TODO если папка не активная, то вложенные блюда создаются в корневой папке SectionID = null
//!!!!!TODO добавить проверку при создании заказа что заказ уже есть в DB и что то подобное

//todo научиться обрабатывать паники + при создании заказа
//todo при не удачном обновлении объекта в RK7, после создания объекта в WOO нужно его откатить и удалить объект в WOO

func main() {
	logger := logging.GetLogger()
	logger.Info("Start Main")
	v := version.GetVersion()
	logger.Infof("Version %s", v.String())
	defer logger.Info("End Main")

	check.Check()
	cfg := config.GetConfig()

	go sync.SyncMenuServiceWithRecovered()
	go telegram.BotStart()

	router := httprouter.New()

	router.GET("/", httphandler.HandlerOtherAll)
	router.POST("/webhook/creat_order", httphandler.HandlerWebhookCreateOrder)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.SERVICE.PORT), router))
}

func init() {
	logger := logging.GetLogger()

	logger.Println("Start main init...")
	defer logger.Println("End main init.")
	cfg := config.GetConfig()
	var err error

	_ = wooapi.NewAPI(cfg.WOOCOMMERCE.URL, cfg.WOOCOMMERCE.Key, cfg.WOOCOMMERCE.Secret)

	_, err = rk7api.NewAPI(cfg.RK7.URL, cfg.RK7.User, cfg.RK7.Pass)
	if err != nil {
		logger.Fatal("failed main init; rk7api.NewAPI; ", err)
	}

	_, err = cache.NewCacheMenu()
	if err != nil {
		logger.Error("failed in cache.NewCacheMenu()")
	}

	if database.Exists(database.DB_NAME) != true {
		logger.Info(database.DB_NAME, " not exist")
		err := database.CreateDB(database.DB_NAME)
		if err != nil {
			logger.Fatalf("%s, %v", database.DB_NAME, err)
		}
	} else {
		logger.Info(database.DB_NAME, " exist")
	}
}
