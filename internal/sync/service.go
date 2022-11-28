package sync

import (
	"WooWithRkeeper/internal/cache"
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/internal/database"
	"WooWithRkeeper/internal/rk7api"
	"WooWithRkeeper/internal/telegram"
	"WooWithRkeeper/internal/wooapi"
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

	cfg := config.GetConfig()

	index := 0 //количество перезапусков при панике
	for {
		SyncMenuService()
		index++

		_, err := cache.NewCacheMenu()
		if err != nil {
			logger.Error("failed in cache.NewCacheMenu()")
		}

		_ = wooapi.NewAPI(cfg.WOOCOMMERCE.URL, cfg.WOOCOMMERCE.Key, cfg.WOOCOMMERCE.Secret)

		_, err = rk7api.NewAPI(cfg.RK7.URL, cfg.RK7.User, cfg.RK7.Pass, "REF")
		if err != nil {
			logger.Fatal("failed main init; rk7api.NewAPI; ", err)
		}

		_, err = rk7api.NewAPI(cfg.RK7MID.URL, cfg.RK7MID.User, cfg.RK7MID.Pass, "MID")
		if err != nil {
			logger.Fatal("failed main init; rk7api.NewAPI; ", err)
		}

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
	RK7API := rk7api.GetAPI("REF")

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

	timeStart := time.Now()
	err = m.RefreshMenu()
	if err != nil {
		logger.Errorf("failed RefreshMenu(); %v", err)
		telegram.SendMessageToTelegramWithLogError(err.Error())
	}
	timeRefreshMenu := time.Now().Sub(timeStart)
	for {
		var timeUpdateSyncCateglist, timeUpdateSyncMenuitems, timeUpdateSyncImages time.Duration

		if cfg.MENUSYNC.SyncCateglist == 1 {
			timeStartSyncCateglist := time.Now()
			verifyVersionResult, err := VerifyVersion(RK7API, DB, "Categlist")
			if err != nil {
				telegram.SendMessageToTelegramWithLogError(fmt.Sprintf("Не удалось выполнить проверку меню. Ошибка при проверке VerifyVersion: %v", err))
			} else if verifyVersionResult {
				logger.Info("Версия справочников Categlist совпадает между RK и DB")
				logger.Info("Проверка не требуется")
			} else {
				logger.Println("Требуется обновление Categlist")
				err := SyncCateglist()
				if err != nil {
					if err.Error() == SYNC_CATEGLIST_NEED_UPDATE {
						logger.Warning("Повторно обновляем, т.к. при создании папок совпало наименование папок")
						err := SyncCateglist()
						if err != nil {
							logger.Error("Ошибка после повторной синхронизации")
							telegram.SendMessageToTelegramWithLogError(fmt.Sprintf("Ошибка после повторной синхронизации Categlist SyncMenu: \n%v\n", err))
						}
					} else {
						telegram.SendMessageToTelegramWithLogError(fmt.Sprintf("Ошибка при синхронизации Categlist SyncMenu: \n%v\n", err))
					}
				} else {
					logger.Infof("Синхронизация Categlist выполнена успешно. Версия справочника Categlist в DB обновлена")
				}
			}
			timeUpdateSyncCateglist = time.Now().Sub(timeStartSyncCateglist)
			logger.Infof("Время обновления Categlist: %s", timeUpdateSyncCateglist)
		}

		if cfg.MENUSYNC.SyncMenuitems == 1 {
			timeStartSyncMenuitems := time.Now()
			// сверить версию справочника Menuitems
			verifyVersionResultMenuitems, err := VerifyVersion(RK7API, DB, "Menuitems")
			if err != nil {
				telegram.SendMessageToTelegramWithLogError(fmt.Sprintf("Не удалось выполнить проверку меню. Ошибка при проверке VerifyVersion: %v", err))
			}
			// сверить версию справочника Price
			verifyVersionResultPrices, err := VerifyVersion(RK7API, DB, "Prices")
			if err != nil {
				telegram.SendMessageToTelegramWithLogError(fmt.Sprintf("Не удалось выполнить проверку меню. Ошибка при проверке VerifyVersion: %v", err))
			}
			if verifyVersionResultMenuitems && verifyVersionResultPrices {
				logger.Info("Версия справочников Menuitems и Prices совпадает между RK и DB")
				logger.Info("Проверка не требуется")
			} else {
				logger.Println("Требуется обновление меню")
				err := SyncMenuitems()
				if err != nil {
					telegram.SendMessageToTelegramWithLogError(fmt.Sprintf("Ошибка при синхронизации меню SyncMenu: \n%v\n", err))
				} else {
					logger.Infof("Синхронизация меню выполнена успешно. Версия справочника Menuitems и Prices в DB обновлена")
				}
			}
			timeUpdateSyncMenuitems = time.Now().Sub(timeStartSyncMenuitems)
			logger.Infof("Время обновления Menuitems: %s", timeUpdateSyncMenuitems)
		}

		if cfg.MENUSYNC.SyncImages == 1 {
			timeStartSyncImages := time.Now()
			err := SyncImages()
			if err != nil {
				telegram.SendMessageToTelegramWithLogError(fmt.Sprintf("Ошибка при синхронизации картинок SyncMenu: \n%v\n", err))
			} else {
				logger.Infof("Синхронизация картинок выполнена успешно")
			}
			timeUpdateSyncImages = time.Now().Sub(timeStartSyncImages)
			logger.Infof("Время обновления Images: %s", timeUpdateSyncImages)
		}

		logger.Info("Тайминги по обновлениями:")
		logger.Infof("RefreshMenu: %s", timeRefreshMenu)
		if cfg.MENUSYNC.SyncCateglist == 1 {
			logger.Infof("Categlist: %s", timeUpdateSyncCateglist)
		}
		if cfg.MENUSYNC.SyncMenuitems == 1 {
			logger.Infof("Menuitems: %s", timeUpdateSyncMenuitems)
		}
		if cfg.MENUSYNC.SyncImages == 1 {
			logger.Infof("Images: %s", timeUpdateSyncImages)
		}

		logger.Infof("Полное время обновления: %s", time.Now().Sub(timeStart))
		logger.Infof("time sleep %d seconds\n", cfg.MENUSYNC.Timeout)

		time.Sleep(time.Second * time.Duration(cfg.MENUSYNC.Timeout))

		timeStart = time.Now()
		err = m.RefreshMenu()
		if err != nil {
			logger.Errorf("failed RefreshMenu(); %v", err)
			telegram.SendMessageToTelegramWithLogError(err.Error())
		}
		timeRefreshMenu = time.Now().Sub(timeStart)
	}
}

// TODO сделать ручник! на обновление меню
