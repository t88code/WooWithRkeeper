package sync

import (
	"WooWithRkeeper/internal/cache"
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/internal/rk7api"
	modelsRK7API "WooWithRkeeper/internal/rk7api/models"
	"WooWithRkeeper/internal/telegram"
	"WooWithRkeeper/internal/wooapi"
	"WooWithRkeeper/pkg/logging"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
)

const (
	DB_NAME_SQLITE                 = "db.db"
	ERROR_PRODUCT_NOT_FOUND        = "API BX24: error_description: Product is not found; error: " //TODO
	ERROR_PRODUCTSECTION_NOT_FOUND = "API BX24: error_description: Раздел не найден.; error: "    //TODO
)

//func SyncCategList()
func SyncCategList(rk7api rk7api.RK7API, wooapi wooapi.WOOAPI, db *sqlx.DB) error {
	logger := logging.GetLogger()
	logger.Println("Start SyncCategList")
	defer logger.Infof("End SyncCategList")

	var SyncMenuErrors []string

	menu, err := cache.GetMenu()
	if err != nil {
		return errors.Wrap(err, "failed in GetMenu()")
	}

	categlists, err := menu.GetCateglistRK7()
	if err != nil {
		return errors.Wrap(err, "failed in menu.GetCateglistRK7()")
	}

	productsCategoriesWooByID, err := menu.GetProductCategoriesWooByID() // todo если не найден в меню, то найти в базе локальной
	if err != nil {
		return errors.Wrap(err, "failed in menu.GetProductCategoriesWooByID()")
	}

	categlistsByIdent, err := menu.GetCateglistsRK7ByIdent()
	if err != nil {
		return errors.Wrap(err, "failed in menu.GetCateglistsRK7ByIdent()")
	}

	// 1 этап синхронизации - синхронизация папок меню без иерархии
	// пройтись по всем блюдам из меню кипера и сравнить с меню из битрикса24
	// цель - проставить везде правильные названия блюд и идентификаторы папок
	logger.Println("Запущен 1й этап синхронизации - синхронизация папок меню без иерархии")
	for i, categlist := range categlists {
		logger.Infof("Папка RK7: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d", categlist.Name, categlist.ItemIdent, categlist.ID_BX24, categlist.SectionID_BX24, categlist.Status)

		if categlist.Ident == 0 { //пропустить Categlist с Ident = 0
			logger.Infof("Обновление не требуется")
			continue
		}

		if categlist.WOO_ID == 0 {
			logger.Info("Не указан идентификатор WOO_ID у папки")
		} else {
			logger.Info("Идентификатор найден")
			}
		}

		//logger.Debugf("ProductSectionListMapByXMLID(len = %d):", len(menu.ProductSectionListMapByXMLID))
		//for key, value := range menu.ProductSectionListMapByXMLID {
		//	logger.Debugf("key: %s; value: %v", key, value)
		//}

		//if productCategory, found := menu.ProductSectionListMapByXMLID[strconv.Itoa(categlist.Ident)]; found { //TODO
		if productCategory, found := productsCategoriesWooByID[categlist.WOO_ID]; found {
			//TODO поиск по productsCategoriesWooByID[categlist.WOO_ID] не оправдан
			//в случае если папка не будет обновлена в RK7, то categlist.WOO_ID==0
			//чтобы защититься от такой ситуации нужно хранить в локальной папке кеш categlist
			//если обновление папки categlist RK7 не произойдет, то пометить что синхра не прошла
			//в лок базе будет categlist.WooID который нужно будет использовать и который потом нужно будет просинхронить в RK7
			//проверки и всю синхру делать через библиотеку Cache
			logger.Infof("Папка найдена в WOO. Name: %s, RkeeperID: %d, ID: %d, Parent: %d", productCategory.Name, productCategory.RkeeperID, productCategory.ID, productCategory.Parent)

			if categlist.Status != 3 {
				err := wooapi.ProductCategoryDelete(productCategory.ID)
				if err != nil {
					SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось удалить Папку в WOO: Name: %s, RkeeperID: %d, ID: %d, Parent: %d, Error: %v, sync 1", productCategory.Name, productCategory.RkeeperID, productCategory.ID, productCategory.Parent, err))
					logger.Info("Не удалось удалить Папку в WOO")
				} else {

					err := menu.DeleteProductCategoryFromCache(categlist.WOO_ID)
					if err != nil {
						return errors.Wrapf(err, "failed in menu.DeleteProductCategoryFromCache(WOO_ID=%d)", categlist.WOO_ID)
					} else {
						logger.Info("Папка успешно удалена из BX24")
					}
				}
				continue
			}

			// если папка совпадает, то пропускаем
			logger.Info("Приступаем к сравнению:")
			logger.Infof("RK.NAME=%s && WOO.NAME=%s", categlist.Name, productCategory.Name)
			logger.Infof("RK.ItemIdent=%d && WOO.RkeeperID=%d - TODO пока поле не добавили, проверка не работает", categlist.ItemIdent, productCategory.RkeeperID) // TODO пока поле не добавили
			logger.Infof("RK.PARENT=%d && productCategory.Parent=%d", categlist.WOO_PARENT_ID, productCategory.Parent)
			//logger.Infof("RK.PARENT=%d && categlistsByIdent[categlist.MainParentIdent].WooID=%d", categlist.WOO_PARENT_ID, categlistsByIdent[categlist.MainParentIdent].WOO_ID)

			if categlist.Name == productCategory.Name &&
				// categlist.ItemIdent == productCategory.RkeeperID && TODO пока поле не добавили
				categlist.WOO_PARENT_ID == productCategory.Parent {
				//categlist.WOO_PARENT_ID == categlistsByIdent[categlist.MainParentIdent].WOO_ID { //TODO папка совпадает но нужно проверить что иерархия совпадает
				logger.Info("Папка RK7 совпадает с WOO. Обновление в WOO не требуется")
				continue
			}

			logger.Info("Папка не совпадает с WOO. Требуется обновление в WOO")

			/*
				if categlist.WOO_ID == 0 {
					//TODO кажется что родитель не может быть 0, получается папку добавили а в RK7 не обновили параметр
					//TODO акутально если делается поиск через доп поле categlist.RkeeperID, потом обязательно надо использовать когда будет поле
					//TODO либо идти путем сохранения папок Categlist в локальной базе SQLLite и потом проверять - было обновление или нет
					//TODO у каждой папки будет статус синхронизации - если ошибка то не просинхронился
					//TODO пока эта проверка бесмысленная
					logger.Info("Папка в Woo c WOO_ID==0. Требуется обновление папки в RK7")
					var categlistUpdate []*modelsRK7API.Categlist
					menu.CateglistItemInRK7[i].ID_BX24 = productSection_Id
					categlistUpdate = append(categlistUpdate, menu.CateglistItemInRK7[i])
					_, err = rk7api.SetRefDataCateglist(categlistUpdate)
					if err != nil {
						menu.CateglistItemInRK7[i].ID_BX24 = 0
						SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось обновить Папку в RK7: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status: %d, sync: 1, error: %v", categlist.Name, categlist.ItemIdent, categlist.ID_BX24, categlist.SectionID_BX24, categlist.Status, err))
						continue
					} else {
						logger.Infof("Папка успешно обновлена в RK7. menu.CateglistItemInRK7[i].ID_BX24 = %d", productSection_Id)
					}
				}
			*/

			productsCategoriesWooByID[categlist.WOO_ID].Name = categlist.Name
			productsCategoriesWooByID[categlist.WOO_ID].Parent = categlist.WOO_PARENT_ID
			logger.Info("Кеш обновлен")
			productCategoryUpdated, err := wooapi.ProductCategoryUpdate(productsCategoriesWooByID[categlist.WOO_ID])
			if err != nil {
				if err.Error() == ERROR_PRODUCTSECTION_NOT_FOUND {

				} else {
					SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось обновить Папку в Woo: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status: %d, error: %v, sync: 1", categlist.Name, categlist.ItemIdent, categlist.ID_BX24, categlist.SectionID_BX24, categlist.Status, err))
					continue
				}
				return errors.Wrap(err, "failed wooapi.ProductCategoryUpdate()")
			} else {
				logger.Info("Папка успешно обновлена в WOOID")
			}

			err = bx24api.ProductSectionUpdate(productSection_Id,
				modelsBX24API.Name(categlist.Name),
				modelsBX24API.XMLID(fmt.Sprint(categlist.ItemIdent)))
			if err != nil {
				if err.Error() == ERROR_PRODUCTSECTION_NOT_FOUND {
					//ошибка ERROR_PRODUCTSECTION_NOT_FOUND - папка не найдена в BX24
					//следовательно папку необходимо создать, если она Активна(Status==3)
					if categlist.Status != 3 {
						logger.Info("Пропускаем неактивную папку")
						continue
					} else {
						logger.Info("Папка не обновилась в BX24. Требуется ее создать в BX24")
						ProductSectionIDBX24, err := bx24api.ProductSectionAdd(categlist.Name,
							modelsBX24API.XMLID(fmt.Sprint(categlist.ItemIdent)))
						if err != nil {
							SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось создать Папку в BX24: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status: %d, error, %v, sync: 1", categlist.Name, categlist.ItemIdent, categlist.ID_BX24, categlist.SectionID_BX24, categlist.Status, err))
							continue
						} else if ProductSectionIDBX24 == 0 {
							SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось создать Папку в BX24, ProductSectionIDBX24 = 0: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status: %d, sync: 1", categlist.Name, categlist.ItemIdent, categlist.ID_BX24, categlist.SectionID_BX24, categlist.Status))
							continue
						} else {
							logger.Info("Папка успешно создана в BX24")
							logger.Info("Необходимо обновить кеш ProductSectionListMapByID, ProductSectionListMapByXMLID")
							productSectionGet, err := bx24api.ProductSectionGet(ProductSectionIDBX24)
							if err != nil {
								return errors.Wrapf(err, "failed in bx24api.ProductSectionGet(%d)", ProductSectionIDBX24)
							} else {
								menu.ProductSectionListMapByID[productSectionGet.ID] = productSectionGet
								menu.ProductSectionListMapByXMLID[productSectionGet.XMLID] = productSectionGet
								logger.Info("Кеш успешно обновлен")
							}
						}

						//ошибок нет, папка создана
						var categlistUpdate []*modelsRK7API.Categlist
						menu.CateglistItemInRK7[i].ID_BX24 = ProductSectionIDBX24
						categlistUpdate = append(categlistUpdate, menu.CateglistItemInRK7[i])
						_, err = rk7api.SetRefDataCateglist(categlistUpdate)
						if err != nil {
							SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось обновить Папку в RK7: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status: %d, sync: 1, error: %v", categlist.Name, categlist.ItemIdent, categlist.ID_BX24, categlist.SectionID_BX24, categlist.Status, err))
							continue
						} else {
							logger.Info("Папка успешно обновлена в RK7")
						}
					}
				} else {
					SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось обновить Папку в BX24: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status: %d, error: %v, sync: 1", categlist.Name, categlist.ItemIdent, categlist.ID_BX24, categlist.SectionID_BX24, categlist.Status, err))
					continue
				}
			} else {
				logger.Info("Папка успешно обновлена в BX24")
				logger.Info("Необходимо обновить кеш ProductSectionListMapByID, ProductSectionListMapByXMLID")
				menu.ProductSectionListMapByID[strconv.Itoa(productSection_Id)].NAME = categlist.Name
				menu.ProductSectionListMapByID[strconv.Itoa(productSection_Id)].XMLID = fmt.Sprint(categlist.ItemIdent)

			}
		} else {
			logger.Info("Папка не найдена в BX24")
			//папка не найдена в BX24 и если статус не активный в RK7, то пропустить
			if categlist.Status != 3 {
				logger.Info("Папка не активная в RK7, пропускаем")
				continue
			}

			logger.Info("Требуется создать папку в BX24")

			//func ProductSectionAddAndUpdateRK
			ProductSectionIDBX24, err := bx24api.ProductSectionAdd(categlist.Name,
				modelsBX24API.XMLID(strconv.Itoa(categlist.ItemIdent)))
			if err != nil {
				SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось создать Папку в BX24: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status: %d, error: %v, sync: 1", categlist.Name, categlist.ItemIdent, categlist.ID_BX24, categlist.SectionID_BX24, categlist.Status, err))
				continue
			} else if ProductSectionIDBX24 == 0 {
				SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось создать Папку в BX24, ProductSectionIDBX24 = 0: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status: %d, sync: 1", categlist.Name, categlist.ItemIdent, categlist.ID_BX24, categlist.SectionID_BX24, categlist.Status))
				continue
			} else {
				logger.Info("Папка успешно создана в BX24")
				logger.Info("Необходимо обновить кеш ProductSectionListMapByID, ProductSectionListMapByXMLID")
				productSectionGet, err := bx24api.ProductSectionGet(ProductSectionIDBX24)
				if err != nil {
					return errors.Wrapf(err, "failed in bx24api.ProductSectionGet(%d)", ProductSectionIDBX24)
				} else {
					menu.ProductSectionListMapByID[productSectionGet.ID] = productSectionGet
					menu.ProductSectionListMapByXMLID[productSectionGet.XMLID] = productSectionGet
					logger.Info("Кеш успешно обновлен")
				}
			}

			var categlistUpdate []*modelsRK7API.Categlist
			menu.CateglistItemInRK7[i].ID_BX24 = ProductSectionIDBX24
			categlistUpdate = append(categlistUpdate, menu.CateglistItemInRK7[i])
			_, err = rk7api.SetRefDataCateglist(categlistUpdate)
			if err != nil {
				menu.CateglistItemInRK7[i].ID_BX24 = 0
				SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось обновить Папку в RK7: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status: %d, sync: 1, error: %v", categlist.Name, categlist.ItemIdent, categlist.ID_BX24, categlist.SectionID_BX24, categlist.Status, err))
				continue
			} else {
				logger.Info("Папка успешно обновлена в RK7")
			}
			//end func ProductSectionAddAndUpdateRK
		}
	}

	if len(SyncMenuErrors) > 0 {
		logger.Println("1й этап синхронизации завершился с ошибками")
		logger.Println("2й этап синхронизации не будет запущен")
		return errors.New(strings.Join(SyncMenuErrors, "\n"))
	}

	//err = menu.RefreshMenu()
	//if err != nil {
	//	return errors.Wrap(err, "ошибка при обновлении меню")
	//}

	// 2 этап синхронизации - синхронизация иерархии папок меню
	// создать CateglistMap
	logger.Println("Запущен 2й этап синхронизации - синхронизация иерархии папок меню")

	// обновить SectionID_BX24 ProductSection и Categlist
	logger.Println("Запущен процесс обновления SectionID_BX24 в ProductSection и Categlist")
	for i, categlist := range menu.CateglistItemInRK7 {
		logger.Infof("Папка RK7: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status: %d, sync: 2", categlist.Name, categlist.ItemIdent, categlist.ID_BX24, categlist.SectionID_BX24, categlist.Status)
		if categlist.Status != 3 {
			logger.Infof("Обновление не требуется")
			continue
		}

		logger.Info("Требуется обновить SectionID в BX24")

		mainParentIdent := categlist.MainParentIdent
		var sectionId string
		if menu.CateglistMapByIdent[mainParentIdent].ID_BX24 != 0 {
			sectionId = strconv.Itoa(menu.CateglistMapByIdent[mainParentIdent].ID_BX24)
		} else {
			sectionId = ""
		}

		if menu.ProductSectionListMapByXMLID[strconv.Itoa(categlist.ItemIdent)].SECTIONID != sectionId {
			err := bx24api.ProductSectionUpdate(categlist.ID_BX24, modelsBX24API.SectionID(sectionId))
			if err != nil {
				SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось обновить Папку в BX24: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status: %d, error: %v, sync: 2", categlist.Name, categlist.ItemIdent, categlist.ID_BX24, categlist.SectionID_BX24, categlist.Status, err))
				continue
			} else {
				logger.Info("Папка успешно обновлена в BX24")
				logger.Info("Необходимо обновить кеш меню")
				menu.ProductSectionListMapByXMLID[strconv.Itoa(categlist.ItemIdent)].SECTIONID = sectionId
				logger.Info("Кеш меню обновлен успешно")
			}

		} else {
			logger.Info("Обновление папки в BX24 не требуется")
			logger.Info("menu.ProductSectionListMapByXMLID[strconv.Itoa(categlist.ItemIdent)].SECTIONID == menu.CateglistMapByIdent[mainParentIdent].ID_BX24")
		}

		if categlist.SectionID_BX24 != menu.CateglistMapByIdent[mainParentIdent].ID_BX24 {
			logger.Info("Требуется обновить SectionID в RK7")
			var categlistUpdate []*modelsRK7API.Categlist
			menu.CateglistItemInRK7[i].SectionID_BX24 = menu.CateglistMapByIdent[mainParentIdent].ID_BX24
			categlistUpdate = append(categlistUpdate, menu.CateglistItemInRK7[i])

			_, err = rk7api.SetRefDataCateglist(categlistUpdate)
			if err != nil {
				SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось обновить Папку в RK7: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status: %d, sync: 2, error: %v", categlist.Name, categlist.ItemIdent, categlist.ID_BX24, categlist.SectionID_BX24, categlist.Status, err))
				continue
			} else {
				logger.Info("Папка успешно обновлена в RK7")
			}
		} else {
			logger.Info("Обновление папки в RK7 не требуется")
			logger.Info("menu.CateglistItemInRK7[i].SectionID_BX24 == menu.CateglistMapByIdent[categlist.MainParentIdent].ID_BX24")
		}
	}

	if len(SyncMenuErrors) > 0 {
		return errors.New(strings.Join(SyncMenuErrors, "\n"))
	}

	VersionRefName, err := GetVersion(rk7api, "Categlist")
	if err != nil {
		return errors.Wrapf(err, "failed GetVersion(rk7api, %s)", "Categlist")
	}

	err = UpdateVersionInDB(db, "Categlist", VersionRefName)
	if err != nil {
		return errors.Wrapf(err, "failed UpdateVersionInDB(db, %s, %d)", "Categlist", VersionRefName)
	}

	return nil
}

//func SyncMenuitems()
func SyncMenuitems(rk7api rk7api.RK7API, wooapi wooapi.WOOAPI, db *sqlx.DB) error {
	logger := logging.GetLogger()
	logger.Println("Start SyncMenuitems")
	defer logger.Println("End SyncMenuitems")

	var SyncMenuErrors []string

	// PRICETYPES-3="9223372036854775807" == PRICE="0.00"

	menu, err := cache.GetMenu()
	if err != nil {
		return errors.Wrap(err, "failed GetMenu()")
	}

	for i, menuitemItem := range menu.MenuitemItemInRK7 {
		logger.Infof("Блюдо RK7: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status: %d, RK_Price: %d", menuitemItem.Name, menuitemItem.ItemIdent, menuitemItem.ID_BX24, menuitemItem.SectionID_BX24, menuitemItem.Status, menuitemItem.PRICETYPES3)
		if menuitemItem.Ident == 0 { //пропустить блюдо с Ident = 0
			logger.Infof("Обновление не требуется")
			continue
		}

		var pricetype3 string
		if menuitemItem.PRICETYPES3 == 9223372036854775807 || menuitemItem.PRICETYPES3 == 0 {
			pricetype3 = "0.00"
		} else {
			p := fmt.Sprint(menuitemItem.PRICETYPES3)
			pricetype3 = fmt.Sprintf("%s.%s", p[:len(p)-2], p[len(p)-2:])
		}

		var categlistParentBX24Id int = menu.CateglistMapByIdent[menuitemItem.MainParentIdent].ID_BX24

		if product, found := menu.ProductListMapByXMLID[strconv.Itoa(menuitemItem.ItemIdent)]; found {

			logger.Infof("Блюдо найдено в BX24. Name: %s, BX24_XML_ID: %s, BX24_ID: %s, BX24_Sextion_ID: %s, BX24_Active: %s, BX24_Price: %s", product.NAME, product.XMLID, product.ID, product.SECTIONID, product.ACTIVE, product.PRICE)

			if menuitemItem.ItemIdent == 0 {
				logger.Info("Блюдо не определено")
				continue
			}

			//если блюдо удалено в RK7, то удалить в BX24
			if menuitemItem.Status != 3 {
				err := bx24api.ProductDel(menuitemItem.ID_BX24)
				if err != nil {
					SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось удалить Блюдо в BX24: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status: %d, error: %v", menuitemItem.Name, menuitemItem.ItemIdent, menuitemItem.ID_BX24, menuitemItem.SectionID_BX24, menuitemItem.Status, err))
					logger.Infof("Не удалось удалить Блюдо в BX24")
				} else {
					logger.Infof("Блюдо успешно удалено из BX24")
				}
				continue
			}

			var productSectionID string
			if product.SECTIONID == "" {
				productSectionID = "0"
			} else {
				productSectionID = product.SECTIONID
			}

			// categlistParentBX24Id - уже истинный,
			// следовательно, у блюд его нужно проверить и установить истинный
			if menuitemItem.SectionID_BX24 != categlistParentBX24Id {
				logger.Infof("Не сходятся menuitemItem.SectionID_BX24=%d и categlistParentBX24Id=%d", menuitemItem.SectionID_BX24, categlistParentBX24Id)
				var menuitem []*modelsRK7API.MenuitemItem
				menu.MenuitemItemInRK7[i].SectionID_BX24 = categlistParentBX24Id
				menuitemItem.SectionID_BX24 = categlistParentBX24Id
				menuitem = append(menuitem, menu.MenuitemItemInRK7[i])

				_, err = rk7api.SetRefDataMenuitems(menuitem)
				if err != nil {
					SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось обновить Блюдо в RK7: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status: %d, error: %v", menuitemItem.Name, menuitemItem.ItemIdent, menuitemItem.ID_BX24, categlistParentBX24Id, menuitemItem.Status, err))
					continue
				}
				logger.Infof("Блюдо успешно обновлено в RK7. SectionID_BX24: %d", categlistParentBX24Id)
			}

			// если папка совпадает, то пропускаем
			logger.Info("Приступаем к сравнению:")
			logger.Infof("RK.NAME=%s && BX.NAME=%s", menuitemItem.Name, product.NAME)
			logger.Infof(`BX.ACTIVE=%s && "Y"`, product.ACTIVE)
			logger.Infof("RK.ItemIdent=%d && BX.XMLID=%s", menuitemItem.ItemIdent, product.XMLID)
			logger.Infof("categlistParentBX24Id=%d && productSectionID=%s", categlistParentBX24Id, productSectionID)
			logger.Infof("RK.pricetype3=%s && BX.PRICE=%s", pricetype3, product.PRICE)

			if product.NAME == menuitemItem.Name &&
				product.ACTIVE == "Y" &&
				product.XMLID == strconv.Itoa(menuitemItem.ItemIdent) &&
				productSectionID == strconv.Itoa(categlistParentBX24Id) &&
				product.PRICE == pricetype3 {
				logger.Info("Блюдо RK7 совпадает с BX24. Обновление в BX24 не требуется")
				continue
			}

			logger.Info("Блюдо не совпадает с BX24. Требуется обновление в BX24")
			err := bx24api.ProductUpdate(menuitemItem.ID_BX24,
				modelsBX24API.Name(menuitemItem.Name),
				modelsBX24API.Active("Y"),
				modelsBX24API.XMLID(strconv.Itoa(menuitemItem.ItemIdent)),
				modelsBX24API.SectionID(strconv.Itoa(categlistParentBX24Id)),
				modelsBX24API.Price(pricetype3))
			if err != nil {
				if err.Error() != ERROR_PRODUCT_NOT_FOUND {
					SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось обновить Блюдо в BX24: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status: %d, error: %v", menuitemItem.Name, menuitemItem.ItemIdent, menuitemItem.ID_BX24, categlistParentBX24Id, menuitemItem.Status, err))
					continue
				} else {
					logger.Info("Блюдо не обновилось в BX24. Требуется его создать в BX24")
					ProductIDBX24, err := bx24api.ProductAdd(menuitemItem.Name,
						modelsBX24API.Active("Y"),
						modelsBX24API.XMLID(fmt.Sprint(menuitemItem.ItemIdent)),
						modelsBX24API.SectionID(fmt.Sprint(categlistParentBX24Id)),
						modelsBX24API.Price(pricetype3))
					if err != nil {
						SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось создать Блюдо в BX24: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status: %d, error: %v", menuitemItem.Name, menuitemItem.ItemIdent, menuitemItem.ID_BX24, categlistParentBX24Id, menuitemItem.Status, err))
						continue
					} else if ProductIDBX24 == 0 {
						SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось обновить Блюдо в BX24, ProductIDBX24 = 0: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d", menuitemItem.Name, menuitemItem.ItemIdent, menuitemItem.ID_BX24, categlistParentBX24Id, menuitemItem.Status))
						continue
					}

					logger.Info("Блюдо успешно создано в BX24")

					var menuitem []*modelsRK7API.MenuitemItem
					menu.MenuitemItemInRK7[i].ID_BX24 = ProductIDBX24
					menuitem = append(menuitem, menu.MenuitemItemInRK7[i])

					_, err = rk7api.SetRefDataMenuitems(menuitem)
					if err != nil {
						SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось обновить Блюдо в RK7: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status: %d, error: %v", menuitemItem.Name, menuitemItem.ItemIdent, menuitemItem.ID_BX24, categlistParentBX24Id, menuitemItem.Status, err))
						continue
					}
					logger.Info("Блюдо успешно обновлено в RK7")

				}
			} else {
				logger.Info("Блюдо успешно обновлено в BX24")
			}
		} else {

			logger.Info("Блюдо не найдено в BX24")

			if menuitemItem.Status != 3 {
				continue
			}

			logger.Info("Требуется создать блюдо в BX24")
			ProductIDBX24, err := bx24api.ProductAdd(menuitemItem.Name,
				modelsBX24API.Active("Y"),
				modelsBX24API.XMLID(strconv.Itoa(menuitemItem.ItemIdent)),
				modelsBX24API.SectionID(strconv.Itoa(categlistParentBX24Id)),
				modelsBX24API.Price(pricetype3))
			if err != nil {
				SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось создать Блюдо в BX24: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status: %d, error: %v", menuitemItem.Name, menuitemItem.ItemIdent, menuitemItem.ID_BX24, categlistParentBX24Id, menuitemItem.Status, err))
				continue
			} else if ProductIDBX24 == 0 {
				SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось обновить Блюдо в BX24, ProductIDBX24 = 0: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d", menuitemItem.Name, menuitemItem.ItemIdent, menuitemItem.ID_BX24, categlistParentBX24Id, menuitemItem.Status))
				continue
			}

			var menuitem []*modelsRK7API.MenuitemItem
			menu.MenuitemItemInRK7[i].ID_BX24 = ProductIDBX24
			menuitem = append(menuitem, menu.MenuitemItemInRK7[i])
			_, err = rk7api.SetRefDataMenuitems(menuitem)
			if err != nil {
				SyncMenuErrors = append(SyncMenuErrors, fmt.Sprintf("Не удалось обновить Блюдо в RK7: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status: %d, error: %v", menuitemItem.Name, menuitemItem.ItemIdent, menuitemItem.ID_BX24, categlistParentBX24Id, menuitemItem.Status, err))
				continue
			}
			logger.Info("Блюдо успешно обновлено в RK7")
		}
	}

	if len(SyncMenuErrors) > 0 {
		return errors.New(strings.Join(SyncMenuErrors, "\n"))
	}

	VersionRefName, err := GetVersion(rk7api, "Menuitems")
	if err != nil {
		return errors.Wrapf(err, "failed GetVersion(rk7api, %s)", "Menuitems")
	}

	err = UpdateVersionInDB(db, "Menuitems", VersionRefName)
	if err != nil {
		return errors.Wrapf(err, "failed UpdateVersionInDB(db, %s, %d)", "Menuitems", VersionRefName)
	}

	return nil
}

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
	// TODO сделать метод который принудительно все обновляет - меню в битриксе, базу локальную обнуляет

	logger := logging.GetLogger()
	logger.Println("Start Service SyncMenu")
	defer logger.Println("End Service SyncMenu")

	defer func() {
		if r := recover(); r != nil {
			telegram.SendMessageToTelegramWithLogError(fmt.Sprintf("произошла критическая ошибка, синхронизация будет перезапущена, ошибка: %v", r))
		}
	}()

	cfg := config.GetConfig()
	RK7API, err := rk7api.NewAPI(cfg.RK7.URL, cfg.RK7.User, cfg.RK7.Pass)
	if err != nil {
		logger.Error(err)
		return
	}
	BX24API := bx24api.GetAPI()

	if Exists(DB_NAME_SQLITE) != true {
		logger.Info(DB_NAME_SQLITE, " not exist")
		err := CreateDB()
		if err != nil {
			logger.Fatalf("%s, %v", DB_NAME_SQLITE, err)
		}
	} else {
		logger.Info(DB_NAME_SQLITE, " exist")
	}

	db, err := sqlx.Connect("sqlite3", DB_NAME_SQLITE)
	if err != nil {
		logger.Fatalf("failed sqlx.Connect, err: %v", err)
	}
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			logger.Fatalf("failed close sqlx.Connect, err: %v", err)
		}
	}(db)

	menu, err := cache.NewMenu()
	if err != nil {
		telegram.SendMessageToTelegramWithLogError(fmt.Sprintf("Ошибка при попытке получить справочники меню RK и товаров BX24, err: %v", err))
	}

	for {
		timeStart := time.Now()
		if cfg.MENUSYNC.SyncCateglist == 1 {
			// сверить справочники Categlist
			verifyVersionResult, err := VerifyVersion(RK7API, db, "Categlist")
			if err != nil {
				telegram.SendMessageToTelegramWithLogError(fmt.Sprintf("Не удалось выполнить проверку меню. Ошибка при проверке VerifyVersion: %v", err))
			} else if verifyVersionResult {
				logger.Info("Версия справочников Categlist совпадает между RK и DB")
				logger.Info("Проверка не требуется")
			} else {
				logger.Println("Требуется обновление Categlist")
				err = menu.RefreshCateglist()
				if err != nil {
					telegram.SendMessageToTelegramWithLogError(fmt.Sprintf("failed menu.RefreshCateglist(); %v", err))
				} else {
					err = menu.RefreshProductSectionList()
					if err != nil {
						telegram.SendMessageToTelegramWithLogError(fmt.Sprintf("failed menu.RefreshProductSectionList(); %v", err))
					} else {
						timeStart := time.Now()
						err := SyncCategList(RK7API, BX24API, db)
						if err != nil {
							telegram.SendMessageToTelegramWithLogError(fmt.Sprintf("Ошибка при синхронизации Categlist SyncMenu: \n%v\n", err))
						} else {
							logger.Infof("Синхронизация Categlist выполнена успешно. Версия справочника Categlist в DB обновлена")
							logger.Infof("Время обновления Categlist(без обновления кеша): %s", time.Now().Sub(timeStart))
						}
					}
				}
			}
		}

		if cfg.MENUSYNC.SyncMenuitems == 1 {
			// сверить справочники Menuitems
			verifyVersionResult, err := VerifyVersion(RK7API, db, "Menuitems")
			if err != nil {
				telegram.SendMessageToTelegramWithLogError(fmt.Sprintf("Не удалось выполнить проверку меню. Ошибка при проверке VerifyVersion: %v", err))
			} else if verifyVersionResult {
				logger.Info("Версия справочников Menuitems совпадает между RK и DB")
				logger.Info("Проверка не требуется")
			} else {
				logger.Println("Требуется обновление меню")
				err = menu.RefreshCateglist()
				if err != nil {
					telegram.SendMessageToTelegramWithLogError(fmt.Sprintf("failed menu.RefreshCateglist(); %v", err))
				} else {
					err = menu.RefreshMenuitems()
					if err != nil {
						telegram.SendMessageToTelegramWithLogError(fmt.Sprintf("failed menu.RefreshMenuitems(); %v", err))
					} else {
						err = menu.RefreshProductList()
						if err != nil {
							telegram.SendMessageToTelegramWithLogError(fmt.Sprintf("failed menu.RefreshProductList(); %v", err))
						} else {
							timeStart := time.Now()
							err := SyncMenuitems(RK7API, BX24API, db)
							if err != nil {
								telegram.SendMessageToTelegramWithLogError(fmt.Sprintf("Ошибка при синхронизации меню SyncMenu: \n%v\n", err))
							} else {
								logger.Infof("Синхронизация меню выполнена успешно. Версия справочника Menuitems в DB обновлена")
								logger.Infof("Время обновления Menuitems(без обновления кеша): %s", time.Now().Sub(timeStart))
							}
						}
					}
				}
			}
		}

		logger.Infof("Полное время обновления: %s", time.Now().Sub(timeStart))
		logger.Infof("time sleep %d minuts\n", cfg.MENUSYNC.Timeout)

		logger.Infof("Запускаем сверку между Categlist и ProductSection ")
		_, err := verifyCateglistWithProductSection()
		if err != nil {
			logger.Errorf("failed verifyCateglistWithProductSection; %v", err)
		}

		logger.Infof("Запускаем сверку между Menuitems и ProductSection ")
		_, err = verifyMenuitemsWithProducts()
		if err != nil {
			logger.Errorf("failed verifyMenuitemsWithProducts; %v", err)
		}

		time.Sleep(time.Minute * time.Duration(cfg.MENUSYNC.Timeout))
	}
}

func verifyCateglistWithProductSection() (bool, error) {
	logger := logging.GetLogger()
	logger.Println("Start verifyCateglistWithProductSection")
	defer logger.Println("End verifyCateglistWithProductSection")

	menu, err := cache.GetMenu()
	if err != nil {
		return false, errors.New("Ошибка при попытке получить меню из кеша")
	}

	if len(menu.CateglistItemInRK7) == 0 {
		err := menu.RefreshCateglist()
		if err != nil {
			return false, err
		}
	}

	if len(menu.ProductSectionListMapByID) == 0 {
		err := menu.RefreshProductSectionList()
		if err != nil {
			return false, err
		}
	}

	indexCateglistLen := 0
	var report []string
	var categlistsWithXMLIDNull []*modelsRK7API.Categlist
	var categlistsWithoutInProductSectionList []*modelsBX24API.ProductSection
	for i, categlist := range menu.CateglistItemInRK7 {
		logger.Infof("Папка RK7: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d", categlist.Name, categlist.ItemIdent, categlist.ID_BX24, categlist.SectionID_BX24, categlist.Status)

		if categlist.Ident == 0 {
			logger.Infof("Categlist с Ident = 0 пропускаем")
			continue
		}

		if categlist.Status != 3 {
			logger.Infof("Categlist с Status != 3 пропускаем")
			continue
		}

		// папки, без XML_ID
		if categlist.ID_BX24 == 0 {
			categlistsWithXMLIDNull = append(categlistsWithXMLIDNull, menu.CateglistItemInRK7[i])
		}

		// папки, которые не найдены в productList
		if _, found := menu.ProductSectionListMapByXMLID[strconv.Itoa(categlist.ItemIdent)]; !found {
			categlistsWithoutInProductSectionList = append(categlistsWithoutInProductSectionList, menu.ProductSectionListMapByXMLID[strconv.Itoa(categlist.ItemIdent)])
		}

		indexCateglistLen++
	}

	//блюда которые не найдены в productSectionList
	if len(categlistsWithoutInProductSectionList) > 0 {
		text := "Папки, которых нет в BX24, но есть RK7:"
		logger.Info(text)
		report = append(report, text)
		for _, product := range categlistsWithoutInProductSectionList {
			text := fmt.Sprintf("Папка BX24: Name: %s, ID: %s, XML_ID: %s, Section_ID: %s", product.NAME, product.ID, product.XMLID, product.SECTIONID)
			logger.Info(text)
			report = append(report, text)
		}
	}

	// папки, без XML_ID
	if len(categlistsWithXMLIDNull) > 0 {
		text := "Папки в RK7 без ID_BX24:"
		logger.Info(text)
		report = append(report, text)
		for _, categlist := range categlistsWithXMLIDNull {
			text := fmt.Sprintf("Папка RK7: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status %d", categlist.Name, categlist.ItemIdent, categlist.ID_BX24, categlist.SectionID_BX24, categlist.Status)
			logger.Info(text)
			report = append(report, text)
		}
	}

	// проверка на количество
	if indexCateglistLen != len(menu.ProductSectionListMapByID) {
		text := fmt.Sprintf("Количество папок не совпадает: RK7=%d, BX24=%d", indexCateglistLen, len(menu.ProductSectionListMapByID))
		report = append(report, text)
	}

	logger.Infof("Длина CateglistItemInRK7: %d", indexCateglistLen)
	logger.Infof("Длина ProductSectionListMap: %d", len(menu.ProductSectionListMapByID))

	cfg := config.GetConfig()
	if cfg.MENUSYNC.TelegramReport == 1 && len(report) > 0 {
		telegram.SendMessageToTelegramWithLogError(strings.Join(report, "\n"))
	}

	return true, nil
}

func verifyMenuitemsWithProducts() (bool, error) {
	logger := logging.GetLogger()
	logger.Println("Start verifyMenuitemsWithProducts")
	defer logger.Println("End verifyMenuitemsWithProducts")

	menu, err := cache.GetMenu()
	if err != nil {
		return false, errors.New("Ошибка при попытке получить меню из кеша")
	}

	if len(menu.MenuitemItemInRK7) == 0 {
		err := menu.RefreshMenuitems()
		if err != nil {
			return false, err
		}
	}

	if len(menu.ProductListMapByXMLID) == 0 {
		err := menu.RefreshProductList()
		if err != nil {
			return false, err
		}
	}

	//посчитать активные блюда
	indexMenuitemsLen := 0
	var report []string
	var menuitemsWithXMLIDNull []*modelsRK7API.MenuitemItem
	var menuitemsWithoutInProductList []*modelsBX24API.Product
	for i, menuitem := range menu.MenuitemItemInRK7 {
		//logger.Infof("Блюдо RK7: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status: %d, RK_Price: %d", menuitem.Name, menuitem.ItemIdent, menuitem.ID_BX24, menuitem.SectionID_BX24, menuitem.Status, menuitem.PRICETYPES3) TODO включить
		if menuitem.Ident == 0 {
			logger.Debugf("MenuitemItem с Ident = 0 пропускаем")
			continue
		}
		if menuitem.Status != 3 {
			//logger.Debugf("MenuitemItem с Status != 3 пропускаем") TODO включить
			continue
		}
		// блюда, без XML_ID
		if menuitem.ID_BX24 == 0 {
			menuitemsWithXMLIDNull = append(menuitemsWithXMLIDNull, menu.MenuitemItemInRK7[i])
		}
		// блюда, которые не найдены в productSectionList
		if _, found := menu.ProductListMapByXMLID[strconv.Itoa(menuitem.ItemIdent)]; !found {
			menuitemsWithoutInProductList = append(menuitemsWithoutInProductList, menu.ProductListMapByXMLID[strconv.Itoa(menuitem.ItemIdent)])
		}
		indexMenuitemsLen++
	}

	//блюда которые не найдены в productList
	if len(menuitemsWithoutInProductList) > 0 {
		text := "Блюда, которых нет в BX24, но есть RK7:"
		logger.Info(text)
		report = append(report, text)
		for _, product := range menuitemsWithoutInProductList {
			text := fmt.Sprintf("Блюдо BX24: Name: %s, ID: %s, XML_ID: %s, Section_ID: %s, Price: %s", product.NAME, product.ID, product.XMLID, product.SECTIONID, product.PRICE)
			logger.Info(text)
			report = append(report, text)
		}
	}

	//блюда которые без XML_ID
	if len(menuitemsWithXMLIDNull) > 0 {
		text := "Блюда в RK7 без ID_BX24:"
		logger.Info(text)
		report = append(report, text)
		for _, menuitem := range menuitemsWithXMLIDNull {
			text := fmt.Sprintf("Блюдо RK7: Name: %s, RK_ID: %d, RK_BX24_ID: %d, RK_Section_ID: %d, RK7_Status: %d, RK_Price: %d", menuitem.Name, menuitem.ItemIdent, menuitem.ID_BX24, menuitem.SectionID_BX24, menuitem.Status, menuitem.PRICETYPES3)
			logger.Info(text)
			report = append(report, text)
		}
	}

	// проверка на количество
	if indexMenuitemsLen != len(menu.ProductListMapByXMLID) {
		text := fmt.Sprintf("Количество блюд не совпадает: RK7=%d, BX24=%d", indexMenuitemsLen, len(menu.ProductListMapByXMLID))
		report = append(report, text)
	}

	logger.Infof("Длина MenuitemItemInRK7: %d", indexMenuitemsLen)
	logger.Infof("Длина ProductListMap: %d", len(menu.ProductListMapByXMLID))

	cfg := config.GetConfig()
	if cfg.MENUSYNC.TelegramReport == 1 && len(report) > 0 {
		telegram.SendMessageToTelegramWithLogError(strings.Join(report, "\n"))
	}

	return true, nil
}

// TODO сделать ручник! на обновление меню
