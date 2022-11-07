package sync

import (
	"WooWithRkeeper/internal/cache"
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/internal/database"
	"WooWithRkeeper/internal/rk7api"
	"WooWithRkeeper/internal/telegram"
	"WooWithRkeeper/pkg/logging"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

//func SyncMenuServiceWithRecovered()
func SyncMenuServiceWithRecovered() {
	logger := logging.GetLogger()
	logger.Println("Start Service SyncMenuServiceWithRecovered")
	defer logger.Println("End Service SyncMenuServiceWithRecovered")
	index := 0 //количество перезапусков
	for {
		SyncMenuService()
		index++
		if index == 3 {
			break
		}
	}
	telegram.SendMessageToTelegramWithLogError("перезапуск SyncMenuService() прекращен")
}

//func SyncMenuService()
func SyncMenuService() {
	// TODO сделать метод который принудительно все обновляет - меню в WOO, базу локальную обнуляет

	logger := logging.GetLogger()
	logger.Println("Start Service SyncMenu")
	defer logger.Println("End Service SyncMenu")

	defer func() {
		if r := recover(); r != nil {
			telegram.SendMessageToTelegramWithLogError(fmt.Sprintf("произошла критическая ошибка, синхронизация будет перезапущена, ошибка: %v", r))
		}
	}()

	cfg := config.GetConfig()
	RK7API := rk7api.GetAPI()

	//TODO обработка ошибок базы!!
	DB, err := sqlx.Connect("sqlite3", database.DB_NAME)
	if err != nil {
		logger.Fatalf("failed sqlx.Connect; %v", err)
	}
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			logger.Fatalf("failed close sqlx.Connect, err: %v", err)
		}
	}(DB)

	m, err := cache.GetMenu()
	if err != nil {
		telegram.SendMessageToTelegramWithLogError(fmt.Sprintf("Ошибка при попытке получить справочники меню RK и товаров BX24, err: %v", err))
	}

	for {
		timeStart := time.Now()
		if cfg.MENUSYNC.SyncCateglist == 1 {
			// сверить справочники Categlist
			verifyVersionResult, err := VerifyVersion(RK7API, DB, "Categlist")
			if err != nil {
				telegram.SendMessageToTelegramWithLogError(fmt.Sprintf("Не удалось выполнить проверку меню. Ошибка при проверке VerifyVersion: %v", err))
			} else if verifyVersionResult {
				logger.Info("Версия справочников Categlist совпадает между RK и DB")
				logger.Info("Проверка не требуется")
			} else {
				logger.Println("Требуется обновление Categlist")
				timeStart := time.Now()
				err := SyncCateglist()
				if err != nil {
					telegram.SendMessageToTelegramWithLogError(fmt.Sprintf("Ошибка при синхронизации Categlist SyncMenu: \n%v\n", err))
				} else {
					logger.Infof("Синхронизация Categlist выполнена успешно. Версия справочника Categlist в DB обновлена")
					logger.Infof("Время обновления Categlist(без обновления кеша): %s", time.Now().Sub(timeStart))
				}
			}
		}

		if cfg.MENUSYNC.SyncMenuitems == 1 {

			// сверить справочники Menuitems
			verifyVersionResultMenuitems, err := VerifyVersion(RK7API, DB, "Menuitems")
			if err != nil {
				telegram.SendMessageToTelegramWithLogError(fmt.Sprintf("Не удалось выполнить проверку меню. Ошибка при проверке VerifyVersion: %v", err))
				continue
			}
			// сверить справочники Price
			verifyVersionResultPrices, err := VerifyVersion(RK7API, DB, "Prices")
			if err != nil {
				telegram.SendMessageToTelegramWithLogError(fmt.Sprintf("Не удалось выполнить проверку меню. Ошибка при проверке VerifyVersion: %v", err))
				continue
			}

			if verifyVersionResultMenuitems && verifyVersionResultPrices {
				logger.Info("Версия справочников Menuitems и Prices совпадает между RK и DB")
				logger.Info("Проверка не требуется")
			} else {
				logger.Println("Требуется обновление меню")
				timeStart := time.Now()
				err := SyncMenuitems()
				if err != nil {
					telegram.SendMessageToTelegramWithLogError(fmt.Sprintf("Ошибка при синхронизации меню SyncMenu: \n%v\n", err))
				} else {
					logger.Infof("Синхронизация меню выполнена успешно. Версия справочника Menuitems и Prices в DB обновлена")
					logger.Infof("Время обновления Menuitems(без обновления кеша): %s", time.Now().Sub(timeStart))
				}
			}
		}

		logger.Infof("Полное время обновления: %s", time.Now().Sub(timeStart))
		logger.Infof("time sleep %d minuts\n", cfg.MENUSYNC.Timeout)

		time.Sleep(time.Minute * time.Duration(cfg.MENUSYNC.Timeout))

		err = m.RefreshMenu()
		if err != nil {
			logger.Errorf("failed RefreshMenu(); %v", err)
			telegram.SendMessageToTelegramWithLogError(err.Error())
		}
	}
}

// TODO сделать ручник! на обновление меню