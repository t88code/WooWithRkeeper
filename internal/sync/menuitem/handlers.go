package menuitem

import (
	"WooWithRkeeper/internal/cache"
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/internal/rk7api"
	modelsRK7API "WooWithRkeeper/internal/rk7api/models"
	"WooWithRkeeper/internal/telegram"
	"WooWithRkeeper/internal/wooapi"
	modelsWOOAPI "WooWithRkeeper/internal/wooapi/models"
	"WooWithRkeeper/pkg/logging"
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

// HandlerOneStage - 1 стадия синхронизации, сверка между RK7/WOO
func HandlerOneStage(menuitemsSync *[]MenuitemSync) error {
	logger := logging.GetLogger()
	logger.Info("Start HandlerOneStage")
	defer logger.Info("End HandlerOneStage")

	cfg := config.GetConfig()

	logger.Debug("Получаем меню из RK7 и WOO")
	menu, err := cache.GetMenu()
	if err != nil {
		return errors.Wrap(err, "failed in GetMenu")
	}

	menuitems, err := menu.GetMenuitems()
	if err != nil {
		return errors.Wrap(err, "failed in GetMenuitems")
	}

	dishRestsByIdent, err := menu.GetDishRestsByIdent()
	if err != nil {
		return errors.Wrap(err, "failed in GetDishRestsByIdent")
	}

	productsWooByID, err := menu.GetProductsWooByID()
	if err != nil {
		return errors.Wrap(err, "failed in GetProductsWooByID")
	}

	categlistsRK7ByIdent, err := menu.GetCateglistsRK7ByIdent()
	if err != nil {
		return errors.Wrap(err, "failed in GetCateglistsRK7ByIdent")
	}

LoopOneStage:
	for i, menuitem := range menuitems {
		logger.Debug("--------------------------------------")
		logger.Debugf("Menuitem: %s", GetMenuitemNotation(menuitem))

		for _, ignoreIdent := range cfg.RK7.MenuitemIdentIgnore {
			if menuitem.ItemIdent == ignoreIdent {
				logger.Debug("Игнорируем по настройкам конфига")
				*menuitemsSync = append(*menuitemsSync, MenuitemSync{
					MenuitemItem: menuitems[i],
					StatusSync:   IGNORE,
				})
				continue LoopOneStage
			}
		}

		if menuitem.Status == 3 {
			logger.Debug("Блюдо активное")
			if menuitem.PRICETYPES != 9223372036854775807 {
				logger.Debug("Проверяем наличие в стоп-листе")
				dishInStopList := false
				if dishRests, foundInStopList := dishRestsByIdent[menuitem.Ident]; foundInStopList {
					if dishRests.Prohibited == 1 || dishRests.Quantity == 0 {
						dishInStopList = true
					}
				}
				if !dishInStopList {
					if categlistParent, found := categlistsRK7ByIdent[menuitem.MainParentIdent]; found {
						if categlistParent.WOO_SYNC == 1 && categlistParent.Status == 3 && categlistParent.ItemIdent != 0 {
							if menuitem.WOO_ID != 0 {
								if product, found := productsWooByID[menuitem.WOO_ID]; found {
									logger.Debug("Блюдо - найдено в WOO. Сверяем")
									var pricetype3 string
									if menuitem.PRICETYPES == 0 {
										pricetype3 = "0.00"
									} else {
										p := fmt.Sprint(menuitem.PRICETYPES)
										pricetype3 = fmt.Sprintf("%s.%s", p[:len(p)-2], p[len(p)-2:])
									}
									var menuitemName string
									if menuitem.WOO_LONGNAME != "" {
										menuitemName = menuitem.WOO_LONGNAME
									} else {
										menuitemName = menuitem.Name
									}

									var productCategoryID int
									if len(product.Categories) == 1 {
										productCategoryID = product.Categories[0].Id
									}

									logger.Debugf("RK.NAME=%s && WOO.NAME=%s", menuitemName, product.Name)
									logger.Debugf("RK.WOO_ID=%d && WOO.ID=%d", menuitem.WOO_ID, product.ID) // пусть будет лишней
									logger.Debugf("RK.WOO_PARENT_ID!=0")
									logger.Debugf("RK.WOO_PARENT_ID=%d && WOO.Category[0]=%d", menuitem.WOO_PARENT_ID, productCategoryID)
									logger.Debugf("RK.WOO_PARENT_ID=%d && RK.Parent.WOO_ID=%d", menuitem.WOO_PARENT_ID, categlistParent.WOO_ID)
									logger.Debugf("RK.PRICE=%s && WOO.RegularPrice=%s", pricetype3, product.RegularPrice)
									logger.Debugf("WOO.Status=%s && %s", product.Status, WOO_PRODUCT_STATUS_ACTIVE)
									logger.Debugf("WOO.StockStatus==%s && %s", product.StockStatus, WOO_PRODUCT_IN_STOCK)

									if menuitemName == product.Name &&
										menuitem.WOO_ID == product.ID &&
										menuitem.WOO_PARENT_ID != 0 &&
										menuitem.WOO_PARENT_ID == productCategoryID &&
										menuitem.WOO_PARENT_ID == categlistParent.WOO_ID &&
										pricetype3 == product.RegularPrice &&
										product.Status == WOO_PRODUCT_STATUS_ACTIVE &&
										product.StockStatus == WOO_PRODUCT_IN_STOCK {
										logger.Debug("Блюдо - не требуется обновление в WOO")
										*menuitemsSync = append(*menuitemsSync, MenuitemSync{
											MenuitemItem: menuitems[i],
											StatusSync:   NOT_NEED_UPDATE,
										})
									} else {
										logger.Debug("Блюдо - требуется обновление в WOO")
										*menuitemsSync = append(*menuitemsSync, MenuitemSync{
											MenuitemItem: menuitems[i],
											StatusSync:   NEED_UPDATE,
										})
									}
								} else {
									logger.Debug("Блюдо - не найдено в WOO")
									*menuitemsSync = append(*menuitemsSync, MenuitemSync{
										MenuitemItem: menuitems[i],
										StatusSync:   NOT_FOUND_IN_WOO,
									})
								}
							} else {
								logger.Debug("Блюдо - не указан WOO_ID")
								*menuitemsSync = append(*menuitemsSync, MenuitemSync{
									MenuitemItem: menuitems[i],
									StatusSync:   NOT_WOO_ID,
								})
							}
						} else {
							logger.Debug("Блюдо - Parent с выключенной синхронизацией или не активный, или корневой раздел")
							*menuitemsSync = append(*menuitemsSync, MenuitemSync{
								MenuitemItem: menuitems[i],
								StatusSync:   PARENT_SYNC_OFF,
							})
						}
					} else {
						logger.Debug("Блюдо - Parent не найден")
						*menuitemsSync = append(*menuitemsSync, MenuitemSync{
							MenuitemItem: menuitems[i],
							StatusSync:   NOT_PARENT,
						})
					}
				} else {
					logger.Debug("Блюдо в стоп-листе")
					*menuitemsSync = append(*menuitemsSync, MenuitemSync{
						MenuitemItem: menuitems[i],
						StatusSync:   STOP_LIST,
					})
				}
			} else {
				logger.Debug("Блюдо не указана цена")
				*menuitemsSync = append(*menuitemsSync, MenuitemSync{
					MenuitemItem: menuitems[i],
					StatusSync:   NOT_PRICE,
				})
			}
		} else {
			logger.Debug("Блюдо не активное")
			*menuitemsSync = append(*menuitemsSync, MenuitemSync{
				MenuitemItem: menuitems[i],
				StatusSync:   NOT_ACTIVE,
			})
		}
	}
	return nil
} // todo срань с ссылками которая мне не понятна до конца

// HandlerTwoStage - 2 стадия синхронизации, запуск обработчиков по каждому статусу
func HandlerTwoStage(menuitemsSync *[]MenuitemSync) error {
	logger := logging.GetLogger()
	logger.Info("Start HandlerTwoStage")
	defer logger.Info("End HandlerTwoStage")

	var mNulledError []string
	var mCreateError []string
	var mUpdateError []string

	for _, menuitemSync := range *menuitemsSync {
		switch menuitemSync.StatusSync {
		case IGNORE:
			err := MenuitemNulledInWooAndRk7(menuitemSync.MenuitemItem)
			if err != nil {
				errorText := fmt.Sprintf("Блюдо %s; %v", GetMenuitemNotation(menuitemSync.MenuitemItem), err)
				logger.Error(errorText)
				mNulledError = append(mNulledError, errorText)
			}
		case NOT_ACTIVE:
			err := MenuitemNulledInWooAndRk7(menuitemSync.MenuitemItem)
			if err != nil {
				errorText := fmt.Sprintf("Блюдо %s; %v", GetMenuitemNotation(menuitemSync.MenuitemItem), err)
				logger.Error(errorText)
				mNulledError = append(mNulledError, errorText)
			}
		case NOT_PRICE:
			err := MenuitemNulledInWooAndRk7(menuitemSync.MenuitemItem)
			if err != nil {
				errorText := fmt.Sprintf("Блюдо %s; %v", GetMenuitemNotation(menuitemSync.MenuitemItem), err)
				logger.Error(errorText)
				mNulledError = append(mNulledError, errorText)
			}
		case STOP_LIST:
			err := MenuitemNulledInWooAndRk7(menuitemSync.MenuitemItem)
			if err != nil {
				errorText := fmt.Sprintf("Блюдо %s; %v", GetMenuitemNotation(menuitemSync.MenuitemItem), err)
				logger.Error(errorText)
				mNulledError = append(mNulledError, errorText)
			}
		case NOT_PARENT:
			err := MenuitemNulledInWooAndRk7(menuitemSync.MenuitemItem)
			if err != nil {
				errorText := fmt.Sprintf("Блюдо %s; %v", GetMenuitemNotation(menuitemSync.MenuitemItem), err)
				logger.Error(errorText)
				mNulledError = append(mNulledError, errorText)
			}
		case PARENT_SYNC_OFF:
			err := MenuitemNulledInWooAndRk7(menuitemSync.MenuitemItem)
			if err != nil {
				errorText := fmt.Sprintf("Блюдо %s; %v", GetMenuitemNotation(menuitemSync.MenuitemItem), err)
				logger.Error(errorText)
				mNulledError = append(mNulledError, errorText)
			}
		case NOT_WOO_ID:
			err := MenuitemCreateInWooAndUpdateInRk7(menuitemSync.MenuitemItem)
			if err != nil {
				errorText := fmt.Sprintf("Блюдо %s; %v", GetMenuitemNotation(menuitemSync.MenuitemItem), err)
				logger.Error(errorText)
				mCreateError = append(mCreateError, errorText)
			}
		case NOT_FOUND_IN_WOO:
			err := MenuitemCreateInWooAndUpdateInRk7(menuitemSync.MenuitemItem)
			if err != nil {
				errorText := fmt.Sprintf("Блюдо %s; %v", GetMenuitemNotation(menuitemSync.MenuitemItem), err)
				logger.Error(errorText)
				mCreateError = append(mCreateError, errorText)
			}
		case NEED_UPDATE:
			err := MenuitemUpdateInWoo(menuitemSync.MenuitemItem)
			if err != nil {
				errorText := fmt.Sprintf("Блюдо %s; %v", GetMenuitemNotation(menuitemSync.MenuitemItem), err)
				logger.Error(errorText)
				mUpdateError = append(mUpdateError, errorText)
			}
		case NOT_NEED_UPDATE:
		default:
			return errors.New(fmt.Sprintf("failed StatusSync=%s", menuitemSync.StatusSync))
		}
	}

	if len(mNulledError) > 0 {
		mNulledError = append([]string{"<strong>Ошибки при обнулении блюд</strong>"}, mNulledError...)
		telegram.SendMessageToTelegramWithLogError(strings.Join(mNulledError, "\n"))
	}

	if len(mCreateError) > 0 {
		mCreateError = append([]string{"<strong>Ошибки при создании блюд</strong>"}, mCreateError...)
		telegram.SendMessageToTelegramWithLogError(strings.Join(mCreateError, "\n"))
	}

	if len(mUpdateError) > 0 {
		mUpdateError = append([]string{"<strong>Ошибки при обновлении блюд</strong>"}, mUpdateError...)
		telegram.SendMessageToTelegramWithLogError(strings.Join(mUpdateError, "\n"))
	}

	return nil
}

// MenuitemNulledInWooAndRk7 - обнулить
// Используется во 2 стадии синхронизации
func MenuitemNulledInWooAndRk7(menuitem *modelsRK7API.MenuitemItem) error {
	logger := logging.GetLogger()
	logger.Info("Start HandlerOneStage")
	defer logger.Info("End HandlerOneStage")

	err := OutOfStockMenuitemInWoo(menuitem)
	if err != nil {
		return err // todo error ALL
	} else {
		logger.Debug("Блюда успешно установлено в статус \"Нет в наличии\" в WOO")
		err := NulledMenuitemInRK7(menuitem)
		if err != nil {
			return err
		} else {
			logger.Debug("Блюда успешно обнулено в RK7")
			return nil
		}
	}
}

// OutOfStockMenuitemInWoo - перевести статус Нет в наличии в WOO
func OutOfStockMenuitemInWoo(menuitem *modelsRK7API.MenuitemItem) error {
	logger := logging.GetLogger()
	logger.Info("Start OutOfStockMenuitemInWoo")
	defer logger.Info("End OutOfStockMenuitemInWoo")

	menu, err := cache.GetMenu()
	if err != nil {
		return err
	}

	productsWooByID, err := menu.GetProductsWooByID()
	if err != nil {
		return err
	}

	logger.Debug("Пробуем установить для блюдо в WOO/кеше WOO статус Нет в наличии")
	if menuitem.WOO_ID != 0 {
		if product, found := productsWooByID[menuitem.WOO_ID]; found {
			logger.Debugf("Product найден в кеше: %s", GetProductNotification(product))
			if product.StockStatus != WOO_PRODUCT_OUT_OF_STOCK {
				recoveryStockStatus := product.StockStatus
				product.StockStatus = WOO_PRODUCT_OUT_OF_STOCK
				woo := wooapi.GetAPI()
				_, err = woo.ProductUpdate(product)
				if err != nil {
					product.StockStatus = recoveryStockStatus
					return errors.Wrap(err, "Ошибка при обновлении блюда. Кеш восстановлен")
				} else {
					logger.Debug("Блюдо успешно обновлено. Кеш обновлен")
					return nil
				}
			} else {
				logger.Debug("Обновление не требуется. product.StockStatus = WOO_PRODUCT_OUT_OF_STOCK")
				return nil
			}
		} else {
			return errors.New(fmt.Sprintf("Product(id=%d) не найден в кеше WOO", menuitem.WOO_ID))
		}
	} else {
		logger.Debug("menuitem.WOO_ID = 0")
		return nil
	}
}

// NulledMenuitemInRK7 - обнулить блюдо в RK7 - свойства WOO_ID и WOO_PARENT_ID
func NulledMenuitemInRK7(menuitem *modelsRK7API.MenuitemItem) error {
	logger := logging.GetLogger()
	logger.Info("Start NulledMenuitemInRK7")
	defer logger.Info("End NulledMenuitemInRK7")

	logger.Debug("Обнуляем WOO_ID/WOO_PARENT_ID в RK7")
	if menuitem.WOO_ID != 0 && menuitem.WOO_PARENT_ID != 0 {
		var menuitems []*modelsRK7API.MenuitemItem
		recoveryWooID := menuitem.WOO_ID
		recoveryWooParentID := menuitem.WOO_PARENT_ID
		menuitem.WOO_ID = 0
		menuitem.WOO_PARENT_ID = 0
		menuitems = append(menuitems, menuitem)

		rk7 := rk7api.GetAPI("REF")
		_, err := rk7.SetRefDataMenuitems(menuitems)
		if err != nil {
			menuitem.WOO_ID = recoveryWooID
			menuitem.WOO_PARENT_ID = recoveryWooParentID
			return errors.Wrap(err, "Ошибка при обнулении WOO_ID/WOO_PARENT_ID в RK7. Кеш восстановлен")
		} else {
			logger.Debug("Блюдо успешно обнулено")
			return nil
		}
	} else {
		logger.Info("Обнуление не требуется")
		return nil
	}
}

// MenuitemCreateInWooAndUpdateInRk7 - создать
// Используется во 2 стадии синхронизации
func MenuitemCreateInWooAndUpdateInRk7(menuitem *modelsRK7API.MenuitemItem) error {
	logger := logging.GetLogger()
	logger.Info("Start MenuitemCreateInWooAndUpdateInRk7")
	defer logger.Info("End MenuitemCreateInWooAndUpdateInRk7")

	woo := wooapi.GetAPI()
	rk7 := rk7api.GetAPI("REF")
	menu, err := cache.GetMenu()
	if err != nil {
		return errors.Wrap(err, "Ошибка при получении кеша меню")
	}
	categlistsRK7ByIdent, err := menu.GetCateglistsRK7ByIdent()
	if err != nil {
		return err //todo error text
	}

	if categlistParent, found := categlistsRK7ByIdent[menuitem.MainParentIdent]; found {
		product := new(modelsWOOAPI.Product)
		if menuitem.WOO_LONGNAME != "" {
			product.Name = menuitem.WOO_LONGNAME
		} else {
			product.Name = menuitem.Name
		}

		if menuitem.PRICETYPES == 0 {
			product.RegularPrice = "0.00"
		} else {
			p := fmt.Sprint(menuitem.PRICETYPES)
			product.RegularPrice = fmt.Sprintf("%s.%s", p[:len(p)-2], p[len(p)-2:])
		}

		productCategory := new(modelsWOOAPI.Category)
		productCategory.Id = categlistParent.WOO_ID
		product.Categories = append(product.Categories, productCategory)

		product.Status = WOO_PRODUCT_STATUS_ACTIVE

		productCreated, err := woo.ProductAdd(product)
		if err != nil {
			return errors.Wrapf(err, "Ошибка при создании блюда в WOO; ProductAdd(Name=%s, Category=%d, Price=%s, Status=%s)",
				product.Name, productCategory.Id, product.RegularPrice, product.Status)
		} else {
			if productCreated != nil {
				logger.Debugf("Блюдо в WOO создано успешно: Name=%s, ID=%d, Categories[0].Id=%d, Slug=%s, Status=%s, RegularPrice=%s",
					productCreated.Name,
					productCreated.ID,
					productCreated.Categories[0].Id,
					productCreated.Slug,
					productCreated.Status,
					productCreated.RegularPrice)
				logger.Debug("Обновляем кеш WOO")
				err = menu.AddProductToCache(productCreated)
				if err != nil {
					return errors.Wrap(err, "Ошибка при добавление блюда в кеш WOO")
				} else {
					logger.Debug("Обновлен кеш WOO. Обновляем свойства в RK7")
					var menuitems []*modelsRK7API.MenuitemItem
					menuitem.WOO_ID = productCreated.ID
					menuitem.WOO_PARENT_ID = categlistParent.WOO_ID
					menuitems = append(menuitems, menuitem)
					_, err = rk7.SetRefDataMenuitems(menuitems)
					if err != nil {
						menuitem.WOO_ID = 0
						menuitem.WOO_PARENT_ID = 0
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
	} else {
		return errors.New(fmt.Sprintf("categlistParent(id=%d) не найден", menuitem.MainParentIdent))
	}
}

// MenuitemUpdateInWoo - обновить
// Используется во 2 стадии синхронизации
func MenuitemUpdateInWoo(menuitem *modelsRK7API.MenuitemItem) error {
	logger := logging.GetLogger()
	logger.Info("Start MenuitemUpdateInWoo")
	defer logger.Info("End MenuitemUpdateInWoo")

	menu, err := cache.GetMenu()
	if err != nil {
		return err
	}
	productsWooByID, err := menu.GetProductsWooByID()
	if err != nil {
		return err
	}
	categlistsRK7ByIdent, err := menu.GetCateglistsRK7ByIdent()
	if err != nil {
		return err //todo error text
	}

	logger.Debug("Пробуем обновить блюдо в WOO")
	if menuitem.WOO_ID != 0 {
		if product, found := productsWooByID[menuitem.WOO_ID]; found {
			logger.Debugf("Product найден в кеше: %s", GetProductNotification(product))
			if categlistParent, found := categlistsRK7ByIdent[menuitem.MainParentIdent]; found {

				recoveryName := product.Name
				if menuitem.WOO_LONGNAME != "" {
					product.Name = menuitem.WOO_LONGNAME
				} else {
					product.Name = menuitem.Name
				}

				recoveryPrice := product.RegularPrice
				if menuitem.PRICETYPES == 0 {
					product.RegularPrice = "0.00"
				} else {
					p := fmt.Sprint(menuitem.PRICETYPES)
					product.RegularPrice = fmt.Sprintf("%s.%s", p[:len(p)-2], p[len(p)-2:])
				}

				recoveryStatus := product.Status
				if product.Status != WOO_PRODUCT_STATUS_ACTIVE {
					product.Status = WOO_PRODUCT_STATUS_ACTIVE
				}

				recoveryStockStatus := product.StockStatus
				if product.StockStatus != WOO_PRODUCT_IN_STOCK {
					product.StockStatus = WOO_PRODUCT_IN_STOCK
				}

				var recoveryCategoryID int
				if len(product.Categories) == 1 {
					recoveryCategoryID = product.Categories[0].Id
				}
				product.Categories = []*modelsWOOAPI.Category{
					{Id: categlistParent.WOO_ID},
				}

				woo := wooapi.GetAPI()
				_, err = woo.ProductUpdate(product)
				if err != nil {
					product.Name = recoveryName
					product.Price = recoveryPrice
					product.Status = recoveryStatus
					product.StockStatus = recoveryStockStatus
					product.Categories = []*modelsWOOAPI.Category{
						{Id: recoveryCategoryID},
					}
					return errors.Wrap(err, "Ошибка при обновлении блюда. Кеш восстановлен")
				} else {
					logger.Debug("Блюдо успешно обновлено. Кеш обновлен")
					return nil
				}
			} else {
				return errors.New(fmt.Sprintf("categlistParent(id=%d) не найден", menuitem.MainParentIdent))
			}
		} else {
			return errors.New(fmt.Sprintf("Product(id=%d) не найден в кеше WOO", menuitem.WOO_ID))
		}
	} else {
		return errors.New("menuitem.WOO_ID = 0")
	}
}

func GetMenuitemNotation(menuitem *modelsRK7API.MenuitemItem) string {
	return fmt.Sprintf("Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d", menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES)
}

func GetProductNotification(product *modelsWOOAPI.Product) string {
	var productCategoryID int
	if len(product.Categories) > 0 {
		productCategoryID = product.Categories[0].Id
	}
	return fmt.Sprintf("Name=%s, ID=%d, Categories[0].Id=%d, Slug=%s, RegularPrice=%s", product.Name, product.ID, productCategoryID, product.Slug, product.RegularPrice)
}
