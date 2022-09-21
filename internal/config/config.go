package config

import (
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
			URL  string
			User string
			Pass string
		}
		RK7MID struct {
			URL           string
			User          string
			Pass          string
			OrderTypeCode int
			TableCode     int
			StationCode   int
			TimeoutError  int
			CurrencyCode  int
		}
		BX24 struct {
			URL            string
			FieldVISITID   string
			FieldORDERNAME string
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
			Timeout int
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
		}
	}
)

var cfg Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
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

		err = gcfg.ReadFileInto(&cfg, "config.ini")
		if err != nil {
			logger.Printf("Config:>Failed to parse gcfg data: %s", err)
		} else {
			logger.Print("Config:>Config is read")
		}
	})

	return &cfg
}
