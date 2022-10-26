package sync

import (
	"WooWithRkeeper/internal/cache"
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/internal/database"
	"WooWithRkeeper/internal/rk7api"
	modelsRK7API "WooWithRkeeper/internal/rk7api/models"
	"WooWithRkeeper/internal/telegram"
	"WooWithRkeeper/internal/wooapi"
	modelsWOOAPI "WooWithRkeeper/internal/wooapi/models"
	"WooWithRkeeper/pkg/logging"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"strings"
)

const WOO_PRODUCT_STATUS_ACTIVE = "publish"

func SyncMenuitems() error {

	logger := logging.GetLogger()
	logger.Info("Start SyncMenuitems")
	defer logger.Info("End SyncMenuitems")

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

	err = menu.RefreshMenuitems()
	if err != nil {
		return err
	}

	err = menu.RefreshProducts()
	if err != nil {
		return err
	}

	err = menu.RefreshCateglist()
	if err != nil {
		return err
	}

	menuitems, err := menu.GetMenuitems()
	if err != nil {
		return err
	}

	menuitemsRK7ByIdent, err := menu.GetMenuitemsRK7ByIdent()
	if err != nil {
		return err
	}

	categlistsRK7ByIdent, err := menu.GetCateglistsRK7ByIdent()
	if err != nil {
		return err
	}

	productsWooByID, err := menu.GetProductsWooByID()
	if err != nil {
		return err
	}

	// активные блюда RK7
	var menuitemsActive int

	// блюда найденные в WOO
	var menuitemsNeedDelInWoo []*modelsRK7API.MenuitemItem            // блюда удаленные в RK7 и найдены в WOO - удалить в WOO, обнулить кеш
	var menuitemsNeedUpdateInWoo []*modelsRK7API.MenuitemItem         // блюда RK7 не совпадают с WOO - обновить в WOO
	var menuitemsNoNeedUpdateInWoo []*modelsRK7API.MenuitemItem       // блюда RK7 совпадают с WOO - ничего не делать
	var menuitemsNotFoundCateglistParent []*modelsRK7API.MenuitemItem // блюда RK7 с неопределенным CateglistParent/WooParentID; Папка RK7 Parent: не найдена в кеше RK7 - сообщить об ошибке

	// блюда не найденные в WOO
	var menuitemsIndefiniteWithWooIDActive []*modelsRK7API.MenuitemItem    // блюда RK7 с не существующим WOO_ID, активные - Обнуляем в кеше и RK7. Создаем в WOO. Обновляем в кеше и RK7.
	var menuitemsIndefiniteWithWooIDNotActive []*modelsRK7API.MenuitemItem // блюда RK7 с не существующим WOO_ID, не активные - Обнуляем в кеше и RK7.
	var menuitemsIndefiniteActive []*modelsRK7API.MenuitemItem             // блюда RK7 не определенные, активные - Обнуляем в кеше и RK7. Создаем в WOO. Обновляем в кеше и RK7.
	var menuitemsIndefiniteNotActive []*modelsRK7API.MenuitemItem          // блюда RK7 не определенные, не активные - Обнуляем в кеше и RK7.

	logger.Info("Запущен процесс синхронизации блюд: свойства Name/WOO_ID/WOO_PARENT_ID/PRICE и проверка на удаленный/активный")

LoopOneStage:
	for i, menuitem := range menuitems {
		logger.Debugf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d", menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES3)
		for _, ignoreIdent := range cfg.RK7MID.MenuitemIdentIgnore {
			if menuitem.ItemIdent == ignoreIdent {
				logger.Debug("Игнорируем по настройкам конфига")
				logger.Debug("--------------------------------------")
				continue LoopOneStage
			}
		}

		if menuitem.Status == 3 {
			menuitemsActive++
		}

		//поиск по WOO_ID
		if menuitem.WOO_ID != 0 {
			logger.Debug("Блюдо с не пустым WOO_ID. Пробуем найти в WOO")
			if product, found := productsWooByID[menuitem.WOO_ID]; found {
				if menuitem.Status != 3 {
					logger.Debugf("Блюдо найдено в WOO. Name: %s, WOO_ID: %d, WOO_ParentId: %d, WOO_Status: %s, WOO_Price: %s", product.Name, product.ID, product.Categories[0].Id, product.Status, product.RegularPrice)
				}

				if menuitem.Status != 3 || menuitem.PRICETYPES3 == 9223372036854775807 {
					logger.Debug("Блюдо не активно или в не указанной ценой в RK7. Необходимо удалить в WOO")
					menuitemsNeedDelInWoo = append(menuitemsNeedDelInWoo, menuitems[i])
				} else {
					logger.Debug("Блюдо активно в RK7. Необходимо сравнить с WOO(свойство Name/WOO_ID/WOO_PARENT_ID/Status/RegularPrice)")

					if categlistParent, found := categlistsRK7ByIdent[menuitem.MainParentIdent]; found {
						logger.Debug("Папка Parent найдена в кеше RK7")

						var parentID int
						if categlistParent.ItemIdent == 0 {
							logger.Debug("Папка Parent корневая - используем WOO_ID из cfg.WOOCOMMERCE.MenuCategoryId")
							parentID = cfg.WOOCOMMERCE.MenuCategoryId
						} else {
							logger.Debug("Папка Parent не корневая - используем WOO_ID из categlistsRK7ByIdent[menuitem.MainParentIdent]")
							parentID = categlistParent.WOO_ID
						}

						var pricetype3 string
						if menuitem.PRICETYPES3 == 0 {
							pricetype3 = "0.00"
						} else {
							p := fmt.Sprint(menuitem.PRICETYPES3)
							pricetype3 = fmt.Sprintf("%s.%s", p[:len(p)-2], p[len(p)-2:])
						}

						var menuitemName string
						if menuitem.WOO_LONGNAME != "" {
							menuitemName = menuitem.WOO_LONGNAME
						} else {
							menuitemName = menuitem.Name
						}

						logger.Debug("Блюдо активно в RK7. Необходимо сравнить с WOO(свойства Name/WOO_ID/WOO_PARENT_ID)")
						logger.Debugf("RK.NAME=%s && WOO.NAME=%s", menuitemName, product.Name)
						logger.Debugf("RK.WOO_ID=%d && WOO.ID=%d", menuitem.WOO_ID, product.ID) // пусть будет лишней
						logger.Debugf("RK.WOO_PARENT_ID=%d && WOO.ParentId=%d", menuitem.WOO_PARENT_ID, product.Categories[0].Id)
						logger.Debugf("RK.WOO_PARENT_ID=%d && RK.Parent.WOO_ID=%d", menuitem.WOO_PARENT_ID, parentID)
						logger.Debugf("RK.PRICE=%s && WOO.RegularPrice=%s", pricetype3, product.RegularPrice)
						logger.Debugf("WOO.Status==%s", WOO_PRODUCT_STATUS_ACTIVE)

						if menuitemName == product.Name &&
							menuitem.WOO_ID == product.ID &&
							menuitem.WOO_PARENT_ID == product.Categories[0].Id &&
							menuitem.WOO_PARENT_ID == parentID &&
							pricetype3 == product.RegularPrice &&
							product.Status == WOO_PRODUCT_STATUS_ACTIVE {
							logger.Debug("Блюдо RK7 совпадает с WOO(свойства Name/WOO_ID/WOO_PARENT_ID/RegularPrice/Status). Обновление в WOO не требуется")
							menuitemsNoNeedUpdateInWoo = append(menuitemsNoNeedUpdateInWoo, menuitems[i])
						} else {
							logger.Debug("Блюдо RK7 не совпадает с WOO(свойства Name/WOO_ID/WOO_PARENT_ID/RegularPrice/Status). Обновление в WOO требуется")
							menuitemsNeedUpdateInWoo = append(menuitemsNeedUpdateInWoo, menuitems[i])
						}
					} else {
						errText = "Папка RK7 Parent не найдена в кеше RK7"
						logger.Error(errText)
						errSync = append(errSync, errText)
						menuitemsNotFoundCateglistParent = append(menuitemsNotFoundCateglistParent, menuitems[i])
					}
				}
			} else {
				if menuitem.Status == 3 && menuitem.PRICETYPES3 != 9223372036854775807 {
					logger.Debug("Блюдо RK7: не найдено в WOO/активная. Обнуляем в кеше и RK7. Создаем в WOO. Обновляем в кеше и RK7.")
					menuitemsIndefiniteWithWooIDActive = append(menuitemsIndefiniteWithWooIDActive, menuitems[i])
				} else {
					logger.Debug("Блюдо RK7: не найдено в WOO/не активная. Обнуляем в кеше и RK7.")
					menuitemsIndefiniteWithWooIDNotActive = append(menuitemsIndefiniteWithWooIDNotActive, menuitems[i])
				}
			}
		} else {
			if menuitem.Status == 3 && menuitem.PRICETYPES3 != 9223372036854775807 {
				logger.Debug("Блюдо RK7: не указано WOO_ID/активная. Обнуляем в кеше и RK7. Создаем в WOO. Обновляем в кеше и RK7.")
				menuitemsIndefiniteActive = append(menuitemsIndefiniteActive, menuitems[i])
			} else {
				logger.Debug("Блюдо RK7: не указано WOO_ID/не активная. Обнуляем в кеше и RK7.")
				menuitemsIndefiniteNotActive = append(menuitemsIndefiniteNotActive, menuitems[i])
			}
		}
		logger.Debug("--------------------------------------")

	}

	logger.Info("Блюда RK7:")
	logger.Infof("Всего: %d", len(menuitems))
	logger.Infof("Активные: %d", menuitemsActive)
	logger.Infof("Игнорировано: %d", len(cfg.RK7MID.MenuitemIdentIgnore))

	logger.Infof("Найдено в WOO: %d", len(menuitemsNeedDelInWoo)+len(menuitemsNeedUpdateInWoo)+len(menuitemsNoNeedUpdateInWoo)+len(menuitemsNotFoundCateglistParent))
	logger.Infof("Удалить в WOO: %d", len(menuitemsNeedDelInWoo))
	logger.Infof("Обновить в WOO: %d", len(menuitemsNeedUpdateInWoo))
	logger.Infof("Совпадают с WOO: %d", len(menuitemsNoNeedUpdateInWoo))
	logger.Infof("Папка RK7 Parent не найдена в кеше RK7(ошибка): %d", len(menuitemsNotFoundCateglistParent))

	logger.Infof("Не найдено в WOO: %d", len(menuitemsIndefiniteWithWooIDActive)+len(menuitemsIndefiniteWithWooIDNotActive)+len(menuitemsIndefiniteActive)+len(menuitemsIndefiniteNotActive))
	logger.Infof("Создать в WOO, WOO_ID неопределен, активная: %d", len(menuitemsIndefiniteWithWooIDActive))
	logger.Infof("Обнулить в RK7, WOO_ID неопределен, не активная: %d", len(menuitemsIndefiniteWithWooIDNotActive))
	logger.Infof("Создать в WOO, папки без WOO_ID, активная: %d", len(menuitemsIndefiniteActive))
	logger.Infof("Обнулить в RK7, папки без WOO_ID, не активная: %d", len(menuitemsIndefiniteNotActive))

	logger.Info("Блюда WOO:")
	logger.Infof("productsWooByID: %d", len(productsWooByID))

	logger.Info("Удаляем блюда в WOO:")
	var delCountInWoo int
	for i, menuitem := range menuitemsNeedDelInWoo {
		logger.Debugf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d", menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES3)
		err = DeleteMenuitemInWoo(menuitemsNeedDelInWoo[i])
		if err != nil {
			logger.Error(err.Error())
			errSync = append(errSync, err.Error())
		} else {
			logger.Info("Блюда успешно удалены в WOO")
			delCountInWoo++
		}
	}
	logger.Infof("Удалено %d блюд в WOO", delCountInWoo)

	logger.Info("Обновляем блюда в WOO и RK7:")
	var updateCountInWoo, updateCountInRK7 int
	for i, menuitem := range menuitemsNeedUpdateInWoo {
		logger.Debugf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d", menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES3)

		err = VerifyAndUpdateParentIDInRK7(menuitemsNeedUpdateInWoo[i])
		if err != nil {
			logger.Error(err.Error())
			errSync = append(errSync, err.Error())
		} else {
			err = UpdateMenuitemInWooAndRK7(menuitemsNeedUpdateInWoo[i])
			if err != nil {
				logger.Error(err.Error())
				errSync = append(errSync, err.Error())
			} else {
				logger.Info("Блюдо успешно обновлено в WOO и RK7")
				updateCountInWoo++
			}
		}
	}
	logger.Infof("Обновлено %d блюд в WOO", updateCountInWoo)
	logger.Infof("Обновлено %d блюд в RK7", updateCountInRK7)

	logger.Info("Создаем блюда в WOO:")
	var createCountInWoo int
	for i, menuitem := range menuitemsIndefiniteWithWooIDActive {
		logger.Debugf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d", menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES3)
		err = VerifyAndUpdateParentIDInRK7(menuitemsIndefiniteWithWooIDActive[i])
		if err != nil {
			logger.Error(err.Error())
			errSync = append(errSync, err.Error())
		} else {
			err = CreateMenuitemInWoo(menuitemsIndefiniteWithWooIDActive[i])
			if err != nil {
				logger.Error(err.Error())
				errSync = append(errSync, err.Error())
			} else {
				logger.Info("Блюда успешно созданы в WOO")
				createCountInWoo++
			}
		}
	}
	for i, menuitem := range menuitemsIndefiniteActive {
		logger.Debugf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d", menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES3)
		err = VerifyAndUpdateParentIDInRK7(menuitemsIndefiniteActive[i])
		if err != nil {
			logger.Error(err.Error())
			errSync = append(errSync, err.Error())
		} else {
			err = CreateMenuitemInWoo(menuitemsIndefiniteActive[i])
			if err != nil {
				logger.Error(err.Error())
				errSync = append(errSync, err.Error())
			} else {
				logger.Info("Блюда успешно созданы в WOO")
				createCountInWoo++
			}
		}
	}
	logger.Infof("Создано %d блюд в WOO", createCountInWoo)

	logger.Info("Обнулить блюда в RK7:")
	var nulledCountInRK7 int
	for i, menuitem := range menuitemsIndefiniteWithWooIDNotActive {
		logger.Debugf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d", menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES3)
		if menuitem.WOO_ID == 0 && menuitem.WOO_PARENT_ID == cfg.WOOCOMMERCE.MenuCategoryId {
			logger.Infof("Обнуление WOO_ID/WOO_PARENT_ID в RK7 не требуется. WOO_ID=0, WOO_PARENT_ID=%d", cfg.WOOCOMMERCE.MenuCategoryId)
		} else {
			err = NulledMenuitemInRK7(menuitemsIndefiniteWithWooIDNotActive[i])
			if err != nil {
				logger.Error(err.Error())
				errSync = append(errSync, err.Error())
			} else {
				logger.Info("Блюда успешно обнулены в RK7")
				nulledCountInRK7++
			}
		}
	}
	for i, menuitem := range menuitemsIndefiniteNotActive {
		logger.Debugf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d", menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES3)
		if menuitem.WOO_ID == 0 && menuitem.WOO_PARENT_ID == cfg.WOOCOMMERCE.MenuCategoryId {
			logger.Infof("Обнуление WOO_ID/WOO_PARENT_ID в RK7 не требуется. WOO_ID=0, WOO_PARENT_ID=%d", cfg.WOOCOMMERCE.MenuCategoryId)
		} else {
			err = NulledMenuitemInRK7(menuitemsIndefiniteNotActive[i])
			if err != nil {
				logger.Error(err.Error())
				errSync = append(errSync, err.Error())
			} else {
				logger.Info("Блюда успешно обнулены в RK7")
				nulledCountInRK7++
			}
		}
	}
	logger.Infof("Обновлены %d блюда в RK7", nulledCountInRK7)

	if len(errSync) > 0 && cfg.MENUSYNC.TelegramReport == 1 {
		logger.Info("Cинхронизация блюд завершилась с ошибками")
		telegram.SendMessageToTelegramWithLogError(strings.Join(errSync, "\n"))
	} else {
		logger.Info("Cинхронизация блюд завершилась успешно")
		VersionRefNameMenuitems, err := GetVersion(rk7API, "Menuitems")
		if err != nil {
			return errors.Wrapf(err, "failed GetVersion(RK7API, %s)", "Menuitems")
		}
		err = UpdateVersionInDB(DB, "Menuitems", VersionRefNameMenuitems)
		if err != nil {
			return errors.Wrapf(err, "failed UpdateVersionInDB(DB, %s, %d)", "Menuitems", VersionRefNameMenuitems)
		}

		VersionRefNamePrices, err := GetVersion(rk7API, "Prices")
		if err != nil {
			return errors.Wrapf(err, "failed GetVersion(RK7API, %s)", "Prices")
		}
		err = UpdateVersionInDB(DB, "Prices", VersionRefNamePrices)
		if err != nil {
			return errors.Wrapf(err, "failed UpdateVersionInDB(DB, %s, %d)", "Prices", VersionRefNamePrices)
		}

		logger.Info("Результат обновление блюд:")
		logger.Infof("Длина categlists(все): %d", len(menuitems))
		logger.Infof("Длина categlistsRK7ByIdent(все): %d", len(menuitemsRK7ByIdent))
		logger.Infof("Длина categoriesWooByID: %d", len(productsWooByID))
	}
	return nil
}

// удалить блюдо из Woo
func DeleteMenuitemInWoo(menuitem *modelsRK7API.MenuitemItem) error {
	//TODO при удалении может быть не найдено блюдо, нужно будет тогда проигнорировать ошибку

	logger := logging.GetLogger()
	logger.Info("Start DeleteMenuitemInWoo")
	defer logger.Info("End DeleteMenuitemInWoo")

	var err error
	woo := wooapi.GetAPI()
	menu, err := cache.GetMenu()
	if err != nil {
		return errors.Wrap(err, "Ошибка при получении кеша меню")
	}
	logger.Infof("Удаляем блюдо из WOO/кеша WOO")
	err = woo.ProductDel(menuitem.WOO_ID)
	if err != nil {
		return errors.Wrap(err, "Ошибка при удалении блюда из WOO")
	} else {
		logger.Info("Обнуляем кеш WOO")
		err = menu.DeleteProductFromCache(menuitem.WOO_ID)
		if err != nil {
			return errors.Wrap(err, "Ошибка при удалении меню из кеша WOO")
		} else {
			logger.Info("Обнулен кеш WOO. Папка успешно удалена из WOO.")
			err = NulledMenuitemInRK7(menuitem)
			if err != nil {
				return errors.Wrap(err, "failed NulledMenuitemInRK7(menuitem)")
			} else {
				return nil
			}
		}
	}
}

// обновить блюдо в Woo - свойство Name,Parent,RegularPrice,Status
func UpdateMenuitemInWooAndRK7(menuitem *modelsRK7API.MenuitemItem) error {
	//TODO если при обновлении не найдено блюдо, то необходимо его создать и после обновить папку RK7

	logger := logging.GetLogger()
	logger.Info("Start UpdateMenuitemInWooAndRK7")
	defer logger.Info("End UpdateMenuitemInWooAndRK7")

	var err error

	woo := wooapi.GetAPI()
	if err != nil {
		return errors.Wrap(err, "Ошибка при получении кеша меню")
	}
	logger.Infof("Обновляем блюдо в WOO/кеше WOO")
	product, err := woo.ProductGet(menuitem.WOO_ID)
	if err != nil {
		return errors.Wrapf(err, "Ошибка при получении ProductGet(ID=%d)", menuitem.WOO_ID)
	} else {
		if product != nil {
			logger.Debugf("Product успешно получен: Name=%s, ID=%d, Categories[0].Id=%d, Slug=%s, RegularPrice=%s", product.Name, product.ID, product.Categories[0].Id, product.Slug, product.RegularPrice)

			var menuitemName string
			if menuitem.WOO_LONGNAME != "" {
				menuitemName = menuitem.WOO_LONGNAME
			} else {
				menuitemName = menuitem.Name
			}

			recoveryName := product.Name
			if product.Name != menuitemName {
				product.Name = menuitemName
			}

			recoveryParent := product.Categories[0].Id
			if product.ParentId != menuitem.WOO_PARENT_ID {
				product.ParentId = menuitem.WOO_PARENT_ID
			}

			var pricetype3 string
			if menuitem.PRICETYPES3 == 0 {
				pricetype3 = "0.00"
			} else {
				p := fmt.Sprint(menuitem.PRICETYPES3)
				pricetype3 = fmt.Sprintf("%s.%s", p[:len(p)-2], p[len(p)-2:])
			}

			recoveryPrice := product.RegularPrice
			if product.RegularPrice != pricetype3 {
				product.RegularPrice = pricetype3
			}

			recoveryStatus := product.Status
			if product.Status != WOO_PRODUCT_STATUS_ACTIVE {
				product.Status = WOO_PRODUCT_STATUS_ACTIVE
			}

			_, err = woo.ProductUpdate(product)
			if err != nil {
				product.Name = recoveryName
				product.ParentId = recoveryParent
				product.Status = recoveryStatus
				product.RegularPrice = recoveryPrice
				return errors.Wrap(err, "Ошибка при обновлении блюда. Кеш восстановлен")
			} else {
				logger.Info("Блюдо успешно обновлено. Кеш обновлен")
				return nil
			}
		} else {
			return errors.New(fmt.Sprintf("Не удалось получить ProductGet(ID=%d)", menuitem.WOO_ID))
		}
	}
}

// обнулить блюдо в RK7 - свойства WOO_ID и WOO_PARENT_ID
func NulledMenuitemInRK7(menuitem *modelsRK7API.MenuitemItem) error {
	logger := logging.GetLogger()
	logger.Info("Start NulledMenuitemInRK7")
	defer logger.Info("End NulledMenuitemInRK7")
	var err error
	rk7 := rk7api.GetAPI()
	cfg := config.GetConfig()
	logger.Info("Обнуляем WOO_ID/WOO_PARENT_ID в RK7.")

	var menuitems []*modelsRK7API.MenuitemItem
	recoveryWooID := menuitem.WOO_ID
	recoveryWooParentID := menuitem.WOO_PARENT_ID
	menuitem.WOO_ID = 0
	menuitem.WOO_PARENT_ID = cfg.WOOCOMMERCE.MenuCategoryId
	menuitems = append(menuitems, menuitem)
	_, err = rk7.SetRefDataMenuitems(menuitems)
	if err != nil {
		menuitem.WOO_ID = recoveryWooID
		menuitem.WOO_PARENT_ID = recoveryWooParentID
		return errors.Wrap(err, "Ошибка при обнулении WOO_ID/WOO_PARENT_ID в RK7. Кеш восстановлен")
	} else {
		logger.Info("Блюдо успешно обнулено")
		return nil
	}
}

// обновить блюдо в RK7 - свойство WOO_PARENT_ID
func UpdateMenuitemInRK7(menuitem *modelsRK7API.MenuitemItem, parentID int) error {
	logger := logging.GetLogger()
	logger.Info("Start UpdateMenuitemInRK7")
	defer logger.Info("End UpdateMenuitemInRK7")
	var err error
	rk7 := rk7api.GetAPI()
	logger.Info("Обновляем WOO_PARENT_ID в RK7/кеше RK7")

	var menuitems []*modelsRK7API.MenuitemItem
	recoveryWooParentID := menuitem.WOO_PARENT_ID
	menuitem.WOO_PARENT_ID = parentID
	menuitems = append(menuitems, menuitem)
	_, err = rk7.SetRefDataMenuitems(menuitems)
	if err != nil {
		menuitem.WOO_PARENT_ID = recoveryWooParentID
		return errors.Wrapf(err, "Ошибка при обновлении WOO_PARENT_ID в RK7(Name=%s,Longname=%s,ID=%d,WOO_ID=%d,ParentID=%d,newParentID=%d). Кеш восстановлен",
			menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, parentID)
	} else {
		logger.Info("Блюдо успешно обновлено в RK7/кеше RK7")
		return nil
	}
}

// создать блюдо в Woo
func CreateMenuitemInWoo(menuitem *modelsRK7API.MenuitemItem) error {
	//TODO подумать какие ошибки могут быть например если не удастся создать в WOO потому что существует

	logger := logging.GetLogger()
	logger.Info("Start CreateMenuitemInWoo")
	defer logger.Info("End CreateMenuitemInWoo")

	var err error
	woo := wooapi.GetAPI()
	if err != nil {
		return errors.Wrap(err, "Ошибка при получении кеша меню")
	}
	rk7 := rk7api.GetAPI()
	menu, err := cache.GetMenu()
	if err != nil {
		return errors.Wrap(err, "Ошибка при получении кеша меню")
	}

	logger.Infof("Создаем блюдо в WOO/кеше WOO")
	product := new(modelsWOOAPI.Product)

	var menuitemName string
	if menuitem.WOO_LONGNAME != "" {
		menuitemName = menuitem.WOO_LONGNAME
	} else {
		menuitemName = menuitem.Name
	}
	product.Name = menuitemName

	categoryParent := new(modelsWOOAPI.Categories)
	categoryParent.Id = menuitem.WOO_PARENT_ID
	product.Categories = append(product.Categories, categoryParent)

	var pricetype3 string
	if menuitem.PRICETYPES3 == 0 {
		pricetype3 = "0.00"
	} else {
		p := fmt.Sprint(menuitem.PRICETYPES3)
		pricetype3 = fmt.Sprintf("%s.%s", p[:len(p)-2], p[len(p)-2:])
	}
	product.RegularPrice = pricetype3

	product.Status = WOO_PRODUCT_STATUS_ACTIVE
	productCreated, err := woo.ProductAdd(product)
	if err != nil {
		return errors.Wrapf(err, "Ошибка при создании блюда в WOO; ProductAdd(Name=%s)", product.Name)
	} else {
		if productCreated != nil {
			logger.Debugf("Блюдо в WOO создано успешно: Name=%s, ID=%d, Categories[0].Id=%d, Slug=%s, Status=%s, RegularPrice=%s",
				productCreated.Name,
				productCreated.ID,
				productCreated.Categories[0].Id,
				productCreated.Slug,
				productCreated.Status,
				productCreated.RegularPrice)
			logger.Info("Обновляем кеш WOO")
			err = menu.AddProductToCache(productCreated)
			if err != nil {
				return errors.Wrap(err, "Ошибка при добавление блюда в кеш WOO")
			} else {
				logger.Info("Обновлен кеш WOO. Обновляем свойства в RK7")
				var menuitems []*modelsRK7API.MenuitemItem
				menuitem.WOO_ID = productCreated.ID
				menuitems = append(menuitems, menuitem)
				_, err = rk7.SetRefDataMenuitems(menuitems)
				if err != nil {
					menuitem.WOO_ID = 0
					return errors.Wrap(err, "Ошибка при обновлении WOO_ID/WOO_PARENT_ID в RK7. Кеш установлен по умолчанию.")
				} else {
					logger.Info("Блюдо успешно обновлено")
					return nil
				}
			}
		} else {
			return errors.New(fmt.Sprintf("Не удалось создать блюдо в WOO; ProductAdd(Name=%d)", menuitem.Name))
		}
	}
}

// обновление WOO_PARENT_ID в RK7
func VerifyAndUpdateParentIDInRK7(menuitem *modelsRK7API.MenuitemItem) error {
	logger := logging.GetLogger()
	logger.Info("Start GetParentIDWithUpdateInRK7")
	defer logger.Info("End GetParentIDWithUpdateInRK7")
	var errText string

	cfg := config.GetConfig()
	menu, err := cache.GetMenu()
	if err != nil {
		return err
	}
	categlistsRK7ByIdent, err := menu.GetCateglistsRK7ByIdent()
	if err != nil {
		return err
	}

	if categlistParent, found := categlistsRK7ByIdent[menuitem.MainParentIdent]; found {
		logger.Debug("Папка Parent найдена в кеше RK7")

		var parentID int
		if categlistParent.ItemIdent == 0 {
			logger.Debug("Папка Parent корневая - используем WOO_ID из cfg.WOOCOMMERCE.MenuCategoryId")
			parentID = cfg.WOOCOMMERCE.MenuCategoryId
		} else {
			logger.Debug("Папка Parent не корневая - используем WOO_ID из categlistsRK7ByIdent[menuitem.MainParentIdent]")
			parentID = categlistParent.WOO_ID
		}

		if menuitem.WOO_PARENT_ID != parentID {
			logger.Debug("Необходимо обновить у блюда WOO_PARENT_ID в RK7")
			err = UpdateMenuitemInRK7(menuitem, parentID)
			if err != nil {
				return err
			} else {
				logger.Info("Обновление успешно выполнено")
				return nil
			}
		} else {
			logger.Info("Обновление у блюда WOO_PARENT_ID в RK7 не требуется")
			return nil
		}
	} else {
		errText = fmt.Sprintf("Папка RK7 Parent не найдена в кеше RK7. Блюдо RK7: Name: %s, Longname: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d", menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES3)
		return errors.New(errText)
	}
}
