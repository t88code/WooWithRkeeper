package categlist

import (
	"WooWithRkeeper/internal/cache"
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/internal/rk7api"
	"WooWithRkeeper/internal/rk7api/models"
	"WooWithRkeeper/internal/telegram"
	"WooWithRkeeper/internal/wooapi"
	modelsWOOAPI "WooWithRkeeper/internal/wooapi/models"
	"WooWithRkeeper/internal/wooapi/options"
	"WooWithRkeeper/pkg/logging"
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

// HandlerCateglistDbOneStage 2 этап - обработка DB.Categlist
func HandlerCateglistDbOneStage(categlistsSync *[]CateglistSync) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerCateglistDbOneStage")
	defer logger.Debug("End HandlerCateglistDbOneStage")

	categlistsSyncByStatus := make(map[string][]CateglistSync)
	for _, categlistSync := range *categlistsSync {
		categlistsSyncByStatus[categlistSync.StatusSync] = append(categlistsSyncByStatus[categlistSync.StatusSync], categlistSync)
	}

	for status, categlistsSync := range categlistsSyncByStatus {
		switch status {
		case IGNORE:
			err := HandlerIgnore(categlistsSync)
			if err != nil {
				return errors.Wrap(err, "failed in HandlerIgnore")
			}
		case SYNC_OFF:
			err := HandlerSyncOff(categlistsSync)
			if err != nil {
				return errors.Wrap(err, "failed in HandlerSyncOff")
			}
		case NOT_ACTIVE:
			err := HandlerNotActive(categlistsSync)
			if err != nil {
				return errors.Wrap(err, "failed in HandlerNotActive")
			}
		case NOT_WOO_ID:
			err := HandlerNotWooId(categlistsSync)
			if err != nil {
				return errors.Wrap(err, "failed in HandlerNotWooId")
			}
		case NOT_FOUND_IN_WOO:
			err := HandlerNotFoundInWoo(categlistsSync)
			if err != nil {
				return errors.Wrap(err, "failed in HandlerNotFoundInWoo")
			}
		case NEED_UPDATE:
			err := HandlerNeedUpdate(categlistsSync)
			if err != nil {
				return errors.Wrap(err, "failed in HandlerNeedUpdate")
			}
		case NOT_NEED_UPDATE:
			err := HandlerNotNeedUpdate(categlistsSync)
			if err != nil {
				return errors.Wrap(err, "failed in HandlerNotNeedUpdate")
			}
		}
	}

	return nil
}

// CateglistNulledInWooAndRK7 - удалить папку в WOO с предварительным поиском и обнулить в RK7.
// Используется во 2 этапе синхронизации папок
func CateglistNulledInWooAndRK7(categlist *models.Categlist) error {
	logger := logging.GetLogger()
	logger.Debug("Start CateglistNulledInWooAndRK7")
	defer logger.Debug("End CateglistNulledInWooAndRK7")

	menu, err := cache.GetMenu()
	if err != nil {
		return errors.Wrap(err, "failed in cache.GetMenu()")
	}

	categlistsRK7ByIdent, err := menu.GetCateglistsRK7ByIdent()
	if err != nil {
		return errors.Wrap(err, "failed in GetCateglistsRK7ByIdent()")
	}

	productCategoriesWooByID, err := menu.GetProductCategoriesWooByID()
	if err != nil {
		return errors.Wrap(err, "failed in GetProductCategoriesWooByID()")
	}

	woo := wooapi.GetAPI()
	rk7 := rk7api.GetAPI("REF")

	if categlistInRk7, ok := categlistsRK7ByIdent[categlist.Ident]; ok {
		logger.Debugf("Папка %s", GetCateglistDescription(categlistInRk7))
		logger.Debug("Пробуем найти и удалить в WOO")
		if categlist.WOO_ID != 0 {
			if productCategory, ok := productCategoriesWooByID[categlist.WOO_ID]; ok {
				logger.Debugf("ProductCategory %s", GetProductCategoryDescription(productCategory))
				err := woo.ProductCategoryDelete(productCategory.ID, options.Force(true))
				if err != nil {
					return errors.Wrap(err, "Не удалось удалить папку в WOO")
				} else {
					delete(productCategoriesWooByID, productCategory.ID)
					logger.Debug("ProductCategory успешно удален. Кэш очищен")
				}
			} else {
				logger.Debugf("ProductCategory(id=%d) не найден в WOO", categlist.WOO_ID)
			}
		} else {
			logger.Debug("WOO_ID = 0")
		}

		if categlistInRk7.WOO_ID != 0 && categlistInRk7.WOO_PARENT_ID != 0 {
			logger.Debug("Обнуляем в RK7")
			recoveryWOO_ID := categlistInRk7.WOO_ID
			recoveryWOO_PARENT_ID := categlistInRk7.WOO_PARENT_ID
			categlistInRk7.WOO_ID = 0
			categlistInRk7.WOO_PARENT_ID = 0
			var categlistItems []*models.Categlist
			categlistItems = append(categlistItems, categlistInRk7)
			_, err := rk7.SetRefDataCateglist(categlistItems)
			if err != nil {
				categlistInRk7.WOO_ID = recoveryWOO_ID
				categlistInRk7.WOO_PARENT_ID = recoveryWOO_PARENT_ID
				return errors.Wrap(err, "Не удалось обнулить папку в RK7")
			} else {
				logger.Debug("Папка успешно обнулена в RK7")
			}
		} else {
			logger.Debug("Обнуление в RK7 не требуется. WOO_ID/WOO_PARENT_ID = 0")
		}
	} else {
		return errors.Wrapf(err, "Папка не найдена в RK7")
	}

	return nil
}

func HandlerIgnore(categlistsSync []CateglistSync) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerIgnore")
	defer logger.Debug("End HandlerIgnore")

	var m []string
	m = append(m, "<strong>Ошибка обработки папок из игнор-листа</strong>")
	for _, categlist := range categlistsSync {
		err := CateglistNulledInWooAndRK7(categlist.Categlist)
		if err != nil {
			m = append(m, fmt.Sprintf("%s; %s", err.Error(), GetCateglistDescription(categlist.Categlist)))
		}
	}
	if len(m) > 1 {
		telegram.SendMessageToTelegramWithLogError(strings.Join(m, "\n"))
	}

	return nil
}

func HandlerSyncOff(categlistsSync []CateglistSync) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerSyncOff")
	defer logger.Debug("End HandlerSyncOff")

	var m []string
	m = append(m, "<strong>Ошибка обработки папок с выключенной синхронизацией</strong>")
	for _, categlist := range categlistsSync {
		err := CateglistNulledInWooAndRK7(categlist.Categlist)
		if err != nil {
			m = append(m, fmt.Sprintf("%s; %s", err.Error(), GetCateglistDescription(categlist.Categlist)))
		}
	}
	if len(m) > 1 {
		telegram.SendMessageToTelegramWithLogError(strings.Join(m, "\n"))
	}

	return nil
}

func HandlerNotActive(categlistsSync []CateglistSync) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerNotActive")
	defer logger.Debug("End HandlerNotActive")

	var m []string
	m = append(m, "<strong>Ошибка обработки не активных папок</strong>")
	for _, categlist := range categlistsSync {
		err := CateglistNulledInWooAndRK7(categlist.Categlist)
		if err != nil {
			m = append(m, fmt.Sprintf("%s; %s", err.Error(), GetCateglistDescription(categlist.Categlist)))
		}
	}
	if len(m) > 1 {
		telegram.SendMessageToTelegramWithLogError(strings.Join(m, "\n"))
	}

	return nil
}

// CateglistCreateInWooAndRK7 - создать папку в WOO и обновить WOO_ID в RK7.
// Используется во 2 этапе синхронизации папок
func CateglistCreateInWooAndRK7(categlist *models.Categlist) error {
	logger := logging.GetLogger()
	logger.Debug("Start CateglistCreateInWooAndRK7")
	defer logger.Debug("End CateglistCreateInWooAndRK7")

	var err error

	menu, err := cache.GetMenu()
	if err != nil {
		return errors.Wrap(err, "failed in cache.GetMenu()")
	}

	categlistsRK7ByIdent, err := menu.GetCateglistsRK7ByIdent()
	if err != nil {
		return errors.Wrap(err, "failed in GetCateglistsRK7ByIdent()")
	}

	productCategoriesWooByID, err := menu.GetProductCategoriesWooByID()
	if err != nil {
		return errors.Wrap(err, "failed in GetProductCategoriesWooByID()")
	}

	woo := wooapi.GetAPI()
	rk7 := rk7api.GetAPI("REF")
	cfg := config.GetConfig()

	if categlistInRk7, ok := categlistsRK7ByIdent[categlist.Ident]; ok {
		logger.Debugf("Папка %s", GetCateglistDescription(categlistInRk7))

		// создаем в WOO
		// обновляем в RK7

		logger.Infof("Создаем папку в WOO/кеше WOO")
		category := new(modelsWOOAPI.ProductCategory)
		var categlistName string
		if categlistInRk7.WOO_LONGNAME != "" {
			categlistName = categlistInRk7.WOO_LONGNAME
		} else {
			categlistName = categlistInRk7.Name
		}
		category.Name = categlistName
		category.Parent = cfg.WOOCOMMERCE.MenuCategoryId
		categoryCreated, ResourceId, err := woo.ProductCategoryAdd(category)
		if err != nil {
			if err.Error() == "code:term_exists; message:Элемент с указанным именем уже существует у родительского элемента.; status:400; display:; details:;" {

				err := woo.ProductCategoryDelete(ResourceId, options.Force(true))
				if err != nil {
					return errors.Wrapf(err, "Не удалось удалить папку в WOO: %s", GetProductCategoryDescription(productCategoriesWooByID[ResourceId]))
				} else {
					delete(productCategoriesWooByID, ResourceId)
					logger.Debugf("ProductCategory(%d) успешно удален. Кэш очищен", ResourceId)
				}
				category.Name = categlistName
				categoryCreated, ResourceId, err = woo.ProductCategoryAdd(category)
				if err != nil {
					return errors.Wrapf(err, "failed in ProductCategoryAdd()")
				}
			} else {
				return errors.Wrapf(err, "failed in ProductCategoryAdd()")
			}
		}

		if categoryCreated != nil {
			logger.Debugf("Папка в WOO создана успешно: Name=%s, ID=%d, Parent=%d, Slug=%s",
				categoryCreated.Name,
				categoryCreated.ID,
				categoryCreated.Parent,
				categoryCreated.Slug)
			logger.Debug("Обновляем кеш WOO")
			err = menu.AddProductCategoryToCache(categoryCreated)
			if err != nil {
				return errors.Wrapf(err, "Ошибка при добавление папки в кеш WOO")
			} else {
				logger.Debug("Обновлен кеш WOO. Обновляем свойства в RK7")
				categlistInRk7.WOO_ID = categoryCreated.ID
				categlistInRk7.WOO_PARENT_ID = cfg.WOOCOMMERCE.MenuCategoryId
				var categlists []*models.Categlist
				categlists = append(categlists, categlistInRk7)
				_, err = rk7.SetRefDataCateglist(categlists)
				if err != nil {
					categlistInRk7.WOO_ID = 0
					categlistInRk7.WOO_PARENT_ID = cfg.WOOCOMMERCE.MenuCategoryId
					return errors.Wrapf(err, "Ошибка при обновлении WOO_ID/WOO_PARENT_ID в RK7. Кеш установлен по умолчанию")
				} else {
					logger.Debug("Папка успешно обновлена")
				}
			}
		} else {
			return errors.Wrapf(err, "Не удалось создать папку в WOO; ProductCategoryAdd(Name=%s)", categlist.Name)
		}
	} else {
		return errors.New("Папка не найдена в RK7")
	}

	return nil
}

func HandlerNotWooId(categlistsSync []CateglistSync) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerNotWooId")
	defer logger.Debug("End HandlerNotWooId")

	var m []string
	m = append(m, "<strong>Ошибка обработки папок без указанного WOO_ID</strong>")
	for _, categlist := range categlistsSync {
		err := CateglistCreateInWooAndRK7(categlist.Categlist)
		if err != nil {
			m = append(m, err.Error())
		}
	}
	if len(m) > 1 {
		telegram.SendMessageToTelegramWithLogError(strings.Join(m, "\n"))
	}

	return nil
}

func HandlerNotFoundInWoo(categlistsSync []CateglistSync) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerNotWooId")
	defer logger.Debug("End HandlerNotWooId")

	var m []string
	m = append(m, "<strong>Ошибка обработки папок, не найденных в WOO</strong>")
	for _, categlist := range categlistsSync {
		err := CateglistCreateInWooAndRK7(categlist.Categlist)
		if err != nil {
			m = append(m, err.Error())
		}
	}
	if len(m) > 1 {
		telegram.SendMessageToTelegramWithLogError(strings.Join(m, "\n"))
	}

	return nil
}

// CateglistUpdateInWoo - обновить папку в WOO и обновить WOO_ID в RK7.
// Используется во 2 этапе синхронизации папок
func CateglistUpdateInWoo(categlist *models.Categlist) error {
	logger := logging.GetLogger()
	logger.Debug("Start CateglistUpdateInWoo")
	defer logger.Debug("End CateglistUpdateInWoo")

	var err error

	menu, err := cache.GetMenu()
	if err != nil {
		return errors.Wrap(err, "failed in cache.GetMenu()")
	}

	productCategoriesWooByID, err := menu.GetProductCategoriesWooByID()
	if err != nil {
		return errors.Wrap(err, "failed in GetProductCategoriesWooByID()")
	}

	woo := wooapi.GetAPI()

	logger.Debugf("Обновляем папку в WOO/кеше WOO")

	if categlist.WOO_ID != 0 {
		if productCategory, found := productCategoriesWooByID[categlist.WOO_ID]; found {
			logger.Debugf("ProductCategory успешно найден: %s", GetProductCategoryDescription(productCategory))

			var categlistName string
			if categlist.WOO_LONGNAME != "" {
				categlistName = categlist.WOO_LONGNAME
			} else {
				categlistName = categlist.Name
			}

			recoveryName := productCategory.Name
			if productCategory.Name != categlistName {
				productCategory.Name = categlistName
			}

			_, err = woo.ProductCategoryUpdate(productCategory)
			if err != nil {
				productCategory.Name = recoveryName
				return errors.Wrapf(err, "failed in ProductCategoryUpdate(ID=%d, Name=%s)", productCategory.ID, productCategory.Name)
			} else {
				logger.Debug("Папка успешно обновлена. Кеш обновлен")
				return nil
			}
		} else {
			return errors.New("Не найден в WOO")
		}
	} else {
		return errors.New("WOO_ID = 0")
	}
}

func HandlerNeedUpdate(categlistsSync []CateglistSync) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerNeedUpdate")
	defer logger.Debug("End HandlerNeedUpdate")

	var m []string
	m = append(m, "<strong>Ошибка обработки при обновлении папок</strong>")
	for _, categlist := range categlistsSync {
		err := CateglistUpdateInWoo(categlist.Categlist)
		if err != nil {
			m = append(m, err.Error())
		}
	}
	if len(m) > 1 {
		telegram.SendMessageToTelegramWithLogError(strings.Join(m, "\n"))
	}

	return nil
}

func HandlerNotNeedUpdate(categlistsSync []CateglistSync) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerNotNeedUpdate")
	defer logger.Debug("End HandlerNotNeedUpdate")

	//"<strong>Ошибка обработки папок, не требующих обновление/strong>"

	return nil
}
