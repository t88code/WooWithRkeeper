package config

import (
	check "WooWithRkeeper/internal/license"
	"fmt"
	"gopkg.in/gcfg.v1"
	"io"
	"log"
	"os"
	"sync"
)

type (
	Config struct {
		RK7 struct {
			URL                      string
			User                     string
			Pass                     string
			CateglistIdentIgnore     []int
			MenuitemIdentIgnore      []int
			CLASSIFICATORGROUPS      int
			CLASSIFICATORGROUPSALLOW int
			CLASSIFICATORGROUPSDENY  int
			PRICETYPE                int
		}
		RK7MID struct {
			URL           string
			User          string
			Pass          string
			OrderTypeCode int
			TableCode     int
			StationCode   int
			TimeoutError  int
			CurrencyCode1 int
			CurrencyCode2 int
		}
		TELEGRAM struct {
			BotToken string
			Debug    int
		}
		MAIL struct {
			Address string
		}
		LOG struct {
			Debug int
		}
		MENUSYNC struct {
			Timeout        int
			SyncImages     int
			SyncMenuitems  int
			SyncCateglist  int
			TelegramReport int
		}
		IMAGESYNC struct {
			Path string
		}
		ORDERSYNC struct {
			Timeout int
		}
		WEBHOOK struct {
			URL   string
			Token string
		}
		SERVICE struct {
			PORT int
		}
		DBSQLITE struct {
			DB string
		}
		CACHE struct {
			TimeUpdate int
		}
		WOOCOMMERCE struct {
			URL            string
			Key            string
			Secret         string
			MenuCategoryId int
			RPS            int
		}
		XMLINTERFACE struct {
			Type      int
			UserName  string
			Password  string
			Token     string
			RestCode  int
			ProductID string
			Guid      string
			URL       string
		}
	}
)

var cfg Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		check.Check()
		err := os.MkdirAll("logs", 0770)
		if err != nil {
			fmt.Println(err)
		}

		file, err := os.OpenFile("logs/config.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640)
		if err != nil {
			fmt.Println(err)
		}

		multiWriter := io.MultiWriter(file, os.Stdout)

		logger := log.New(multiWriter, "MAIN ", log.Ldate|log.Ltime|log.Lshortfile)

		logger.Print("Config:>Read application configurations")

		err = gcfg.ReadFileInto(&cfg, "./config/config.ini")
		if err != nil {
			logger.Fatalf("Config:>Failed to parse gcfg data: %s", err)
		} else {
			logger.Print("Config:>Config is read")
		}
	})

	return &cfg
}
