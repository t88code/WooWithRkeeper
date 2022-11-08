package sync

import (
	"WooWithRkeeper/internal/cache"
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/internal/database"
	"WooWithRkeeper/internal/rk7api"
	"WooWithRkeeper/pkg/logging"
	"fmt"
	"github.com/jmoiron/sqlx"
	"os"
)

func SyncImages() error {

	logger := logging.GetLogger()
	logger.Info("Start SyncImages")
	defer logger.Info("End SyncImages")
	var err error
	var errSync []string
	var errText string
	cfg := config.GetConfig()
	rk7API := rk7api.GetAPI()
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

	logger.Debug("Получаем меню из RK7 и WOO")
	menu, err := cache.GetMenu()
	if err != nil {
		return err
	}

	menuitems, err := menu.GetMenuitems()
	if err != nil {
		return err
	}

	logger.Info("Запущен процесс обновления картинок блюд")

LoopOneStage:
	for i, menuitem := range menuitems {
		dish := fmt.Sprintf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d, RK_Image: %s",
			menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES3, menuitem.WOO_IMAGE)
		logger.Debugf(dish)
		for _, ignoreIdent := range cfg.RK7MID.MenuitemIdentIgnore {
			if menuitem.ItemIdent == ignoreIdent {
				logger.Debug("Игнорируем по настройкам конфига")
				logger.Debug("--------------------------------------")
				continue LoopOneStage
			}
		}

		if menuitem.Status == 3 {
			logger.Info("Не активное блюдо. Пропускаем")
		} else {
			if menuitem.WOO_IMAGE == "" {
				errText = "Не указано наименование картинки. Пропускаем"
				logger.Info("Не указано наименование картинки. Пропускаем")
				errSync = append(errSync, fmt.Sprintf("%s; %s", errText, dish))
			} else {
				logger.Infof("Указано наименование картинки %s", menuitem.WOO_IMAGE)
				logger.Info("Проверяем наличие картинки в папке images")
				//TODO проверить наличие картинки в папке
				if _, err := os.Stat("/path/to/whatever"); !os.IsNotExist(err) {
					logger.Info("Файл картинки существует")
					// название картинки совпадает
					// проверить дату изменения картинки
					// -если дата изменения совпадает между картинкой и DB
					// --ничего не делать
					// -если дата изменения не совпадает между картинкой и DB
					// --обновить картинку в WOO
					// --обновить картинку в DB
				} else {
					// название картинки не совпадает
					logger.Info("Файл картинки не существует")
					//отправить сообщение в телеграм
				}
			}
		}
	}

	return nil
}
