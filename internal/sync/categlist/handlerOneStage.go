package categlist

import (
	"WooWithRkeeper/internal/cache"
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/pkg/logging"
)

// HandlerCateglistToDb 1 этап
func HandlerCateglistToDb(categlistsSync *[]CateglistSync) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerCateglistToDb")
	defer logger.Debug("End HandlerCateglistToDb")

	var err error
	cfg := config.GetConfig()

	logger.Debug("Получаем меню из RK7 и WOO")
	menu, err := cache.GetMenu()
	if err != nil {
		return err
	}

	categlists, err := menu.GetCateglistRK7()
	if err != nil {
		return err
	}

	categoriesWooByID, err := menu.GetProductCategoriesWooByID()
	if err != nil {
		return err
	}

	// папки RK7
	var categlistActive int
	var categlistNotActive int

LoopOneStage:
	for _, categlist := range categlists {
		logger.Debug("--------------------------------------")
		logger.Debugf("Папка RK7: %s", GetCateglistDescription(categlist))

		categlistSync := CateglistSync{Categlist: categlist}

		logger.Debug("Проверка игнор-лист")
		for _, ignoreIdent := range cfg.RK7.CateglistIdentIgnore {
			if categlist.ItemIdent == ignoreIdent {
				logger.Debug("Папка в игнор-листе. Обнулить WOO/RK7")
				categlistSync.StatusSync = IGNORE
				*categlistsSync = append(*categlistsSync, categlistSync)
				continue LoopOneStage
			}
		}

		if categlist.WOO_SYNC == 1 {
			logger.Debug("Синхронизация включена")
			if categlist.Status == 3 {
				logger.Debug("Папка активная")
				categlistActive++
				if categlist.WOO_ID != 0 {
					logger.Debug("Указан WOO_ID")
					if category, found := categoriesWooByID[categlist.WOO_ID]; found {
						logger.Debug("Папка найдена в WOO")
						var categlistName string
						if categlist.WOO_LONGNAME != "" {
							categlistName = categlist.WOO_LONGNAME
						} else {
							categlistName = categlist.Name
						}
						logger.Debugf("RK.NAME=%s && RK.LongName=%s && WOO.NAME=%s", categlist.Name, categlist.WOO_LONGNAME, category.Name)
						if categlistName == category.Name {
							logger.Debug("Папка RK7 совпадает с WOO(свойства Name/LongName). Обновление в WOO не требуется")
							categlistSync.StatusSync = NOT_NEED_UPDATE
							*categlistsSync = append(*categlistsSync, categlistSync)
						} else {
							logger.Debug("Папка RK7 не совпадает с WOO(свойства Name/LongName). Обновляем WOO.")
							categlistSync.StatusSync = NEED_UPDATE
							*categlistsSync = append(*categlistsSync, categlistSync)
						}
					} else {
						logger.Debug("Папка не найдена в WOO. Создаем в WOO.")
						categlistSync.StatusSync = NOT_FOUND_IN_WOO
						*categlistsSync = append(*categlistsSync, categlistSync)
					}
				} else {
					logger.Debug("Не указан WOO_ID. Создаем в WOO.")
					categlistSync.StatusSync = NOT_WOO_ID
					*categlistsSync = append(*categlistsSync, categlistSync)
				}
			} else {
				logger.Debug("Папка не активная. Обнуляем WOO/RK7.")
				categlistNotActive++
				categlistSync.StatusSync = NOT_ACTIVE
				*categlistsSync = append(*categlistsSync, categlistSync)
			}
		} else {
			logger.Debug("Синхронизация отключена. Обнуляем в WOO/RK7.")
			categlistSync.StatusSync = SYNC_OFF
			*categlistsSync = append(*categlistsSync, categlistSync)
		}
	}
	return nil
}
