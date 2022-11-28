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

func SyncCateglist() error {

	logger := logging.GetLogger()
	logger.Info("Start SyncCateglist")
	defer logger.Info("End SyncCateglist")

	var err error
	var resultSyncAll []string
	var resultSyncError []string
	var errText string
	var flagNeedUpdate bool = false

	cfg := config.GetConfig()
	rk7API := rk7api.GetAPI("REF")
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

	categlists, err := menu.GetCateglistRK7()
	if err != nil {
		return err
	}

	categlistsRK7ByIdent, err := menu.GetCateglistsRK7ByIdent()
	if err != nil {
		return err
	}

	categoriesWooByID, err := menu.GetProductCategoriesWooByID()
	if err != nil {
		return err
	}

	categoriesWooBySlug, err := menu.GetProductCategoriesWooBySlug()
	if err != nil {
		return err
	}

	// папки RK7
	var categlistActive int
	// todo добавить неактивные блюда

	// папки найденные в WOO
	var categlistNeedDelInWoo []*modelsRK7API.Categlist      // папки удаленные в RK7 и найдены в WOO - удалить в WOO
	var categlistNeedUpdateInWoo []*modelsRK7API.Categlist   // папки RK7 не совпадают с WOO - обновить в WOO
	var categlistNoNeedUpdateInWoo []*modelsRK7API.Categlist // папки RK7 совпадают с WOO - ничего не делать

	// папки не найденные в WOO
	var categlistIndefiniteWithWooIDActive []*modelsRK7API.Categlist    // папки RK7 с не существующим WOO_ID, активные - Обнуляем в кеше и RK7. Создаем в WOO. Обновляем в кеше и RK7.
	var categlistIndefiniteWithWooIDNotActive []*modelsRK7API.Categlist // папки RK7 с не существующим WOO_ID, не активные - Обнуляем в кеше и RK7.
	var categlistIndefiniteActive []*modelsRK7API.Categlist             // папки RK7 не определенные, активные - Обнуляем в кеше и RK7. Создаем в WOO. Обновляем в кеше и RK7.
	var categlistIndefiniteNotActive []*modelsRK7API.Categlist          // папки RK7 не определенные, не активные - Обнуляем в кеше и RK7.

	logger.Info("Запущен 1-й этап синхронизации: свойства Name/WOO_ID и проверка на удаленный/активный")

LoopOneStage:
	for i, categlist := range categlists {

		logger.Debugf("Папка RK7: Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status)

		for _, ignoreIdent := range cfg.RK7.CateglistIdentIgnore {
			if categlist.Ident == ignoreIdent {
				logger.Debug("Игнорируем по настройкам конфига")
				logger.Debug("--------------------------------------")
				continue LoopOneStage
			}
		}

		if categlist.Status == 3 {
			categlistActive++
		}
		if categlist.WOO_ID != 0 {
			logger.Debug("Папка с не пустым WOO_ID. Пробуем найти в WOO")
			if category, found := categoriesWooByID[categlist.WOO_ID]; found {
				logger.Debug("Папка найдена в WOO")
				if categlist.Status != 3 {
					logger.Debug("Папка не активна в RK7. Необходимо удалить в WOO")
					categlistNeedDelInWoo = append(categlistNeedDelInWoo, categlists[i])
				} else {
					logger.Debug("Папка активна в RK7. Необходимо сравнить с WOO(свойства Name и WOO_ID)")

					var categlistName string
					if categlist.WOO_LONGNAME != "" {
						categlistName = categlist.WOO_LONGNAME
					} else {
						categlistName = categlist.Name
					}

					logger.Debugf("RK.NAME=%s && WOO.NAME=%s", categlistName, category.Name)
					logger.Debugf("RK.WOO_ID=%d && WOO.ID=%d", categlist.WOO_ID, category.ID)
					if categlistName == category.Name && categlist.WOO_ID == category.ID {
						logger.Debug("Папка RK7 совпадает с WOO(свойства Name и WOO_ID). Обновление в WOO не требуется")
						categlistNoNeedUpdateInWoo = append(categlistNoNeedUpdateInWoo, categlists[i])
					} else {
						logger.Debug("Папка RK7 не совпадает с WOO(свойства Name и WOO_ID). Обновление в WOO требуется")
						categlistNeedUpdateInWoo = append(categlistNeedUpdateInWoo, categlists[i])
					}
				}
			} else {
				if categlist.Status == 3 {
					logger.Debug("Папка RK7: не найдена в WOO/активная. Обнуляем в кеше и RK7. Создаем в WOO. Обновляем в кеше и RK7.")
					categlistIndefiniteWithWooIDActive = append(categlistIndefiniteWithWooIDActive, categlists[i])
				} else {
					logger.Debug("Папка RK7: не найдена в WOO/не активная. Обнуляем в кеше и RK7.")
					categlistIndefiniteWithWooIDNotActive = append(categlistIndefiniteWithWooIDNotActive, categlists[i])
				}
			}
		} else {
			if categlist.Status == 3 {
				logger.Debug("Папка RK7: не указано WOO_ID/активная. Обнуляем в кеше и RK7. Создаем в WOO. Обновляем в кеше и RK7.")
				categlistIndefiniteActive = append(categlistIndefiniteActive, categlists[i])
			} else {
				logger.Debug("Папка RK7: не указано WOO_ID/не активная. Обнуляем в кеше и RK7.")
				categlistIndefiniteNotActive = append(categlistIndefiniteNotActive, categlists[i])
			}
		}
		logger.Debug("--------------------------------------")
	}

	logger.Info("Папки RK7:")
	logger.Infof("Всего: %d", len(categlists))
	logger.Infof("Активные: %d", categlistActive)
	logger.Infof("Игнорировано: %d", len(cfg.RK7.CateglistIdentIgnore))

	logger.Infof("Найдено в WOO: %d", len(categlistNeedDelInWoo)+len(categlistNeedUpdateInWoo)+len(categlistNoNeedUpdateInWoo))
	logger.Infof("Удалить в WOO: %d", len(categlistNeedDelInWoo))
	logger.Infof("Обновить в WOO: %d", len(categlistNeedUpdateInWoo))
	logger.Infof("Совпадают с WOO: %d", len(categlistNoNeedUpdateInWoo))

	logger.Infof("Не найдено в WOO: %d", len(categlistIndefiniteWithWooIDActive)+len(categlistIndefiniteWithWooIDNotActive)+len(categlistIndefiniteActive)+len(categlistIndefiniteNotActive))
	logger.Infof("Создать в WOO, WOO_ID неопределен, активная: %d", len(categlistIndefiniteWithWooIDActive))
	logger.Infof("Обнулить в RK7, WOO_ID неопределен, не активная: %d", len(categlistIndefiniteWithWooIDNotActive))
	logger.Infof("Создать в WOO, папки без WOO_ID, активная: %d", len(categlistIndefiniteActive))
	logger.Infof("Обнулить в RK7, папки без WOO_ID, не активная: %d", len(categlistIndefiniteNotActive))

	logger.Info("Папки WOO:")
	logger.Infof("ProductCategoriesWooByID: %d", len(categoriesWooByID))
	logger.Infof("ProductCategoriesWooBySlug: %d", len(categoriesWooBySlug))

	//++++
	if len(categlistNeedDelInWoo) > 0 {
		logger.Info("Удаляем в WOO:")
		var delCountInWoo int
		for i, categlist := range categlistNeedDelInWoo {
			c := fmt.Sprintf("Папка RK7: Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status)
			logger.Debug(c)
			err = DeleteCateglistInWoo(categlistNeedDelInWoo[i])
			if err != nil {
				errText = fmt.Sprintf("%s;%v", c, err)
				logger.Error(err.Error())
				resultSyncAll = append(resultSyncAll, errText)
				resultSyncError = append(resultSyncAll, errText)
			} else {
				text := fmt.Sprintf("%s, папка успешно удалена в WOO", c)
				logger.Debug(text)
				resultSyncAll = append(resultSyncAll, text)
				delCountInWoo++
			}
		}
		logger.Infof("Удалено %d папок в WOO", delCountInWoo)
	}

	//++++
	if len(categlistNeedUpdateInWoo) > 0 {
		logger.Info("Обновляем в WOO:")
		var updateCountInWoo int
		for i, categlist := range categlistNeedUpdateInWoo {
			c := fmt.Sprintf("Папка RK7: Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status)
			logger.Debug(c)
			err = UpdateCateglistInWoo(categlistNeedUpdateInWoo[i])
			if err != nil {
				errText = fmt.Sprintf("%s;%v", c, err)
				logger.Error(err.Error())
				resultSyncAll = append(resultSyncAll, errText)
				resultSyncError = append(resultSyncAll, errText)
			} else {
				text := fmt.Sprintf("%s, папка успешно обновлена в WOO", c)
				logger.Debug(text)
				resultSyncAll = append(resultSyncAll, text)
				updateCountInWoo++
			}
		}
		logger.Infof("Обновлено %d папок в WOO", updateCountInWoo)
	}

	//++++
	if len(categlistIndefiniteWithWooIDActive) > 0 ||
		len(categlistIndefiniteActive) > 0 {
		logger.Info("Создаем в WOO:")
		var createCountInWoo int
		for i, categlist := range categlistIndefiniteWithWooIDActive {
			c := fmt.Sprintf("Папка RK7: Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status)
			logger.Debug(c)
			err := CreateCateglistInWoo(categlistIndefiniteWithWooIDActive[i])
			if err != nil {
				errText = fmt.Sprintf("%s;%v", c, err)
				logger.Error(err.Error())
				resultSyncAll = append(resultSyncAll, errText)
				if errors.Unwrap(errors.Unwrap(err)).Error() == ERROR_CREATE_PRODUCTCATEGORY_EXIST {
					originalName := categlistIndefiniteWithWooIDActive[i].Name
					categlistIndefiniteWithWooIDActive[i].Name = fmt.Sprintf("%s_%d", categlistIndefiniteWithWooIDActive[i].Name, categlistIndefiniteWithWooIDActive[i].Code)
					text := fmt.Sprintf("Элемент с указанным именем уже существует у родительского элемента. Пробуем повторно создать с именем: %s", categlistIndefiniteWithWooIDActive[i].Name)
					logger.Warning(text)
					resultSyncAll = append(resultSyncAll, text)
					err := CreateCateglistInWoo(categlistIndefiniteWithWooIDActive[i])
					if err != nil {
						errText = fmt.Sprintf("%s;%v", c, err)
						logger.Error(err.Error())
						resultSyncAll = append(resultSyncAll, errText)
						resultSyncError = append(resultSyncAll, errText)
					} else {
						text := "Папка успешно создана"
						logger.Debug(text)
						resultSyncAll = append(resultSyncAll, text)
						flagNeedUpdate = true
					}
					categlistIndefiniteWithWooIDActive[i].Name = originalName
				} else {
					resultSyncError = append(resultSyncAll, errText)
				}
			} else {
				text := fmt.Sprintf("%s, папка успешно создана в WOO", c)
				logger.Debug(text)
				resultSyncAll = append(resultSyncAll, text)
				createCountInWoo++
			}
		}
		for i, categlist := range categlistIndefiniteActive {
			c := fmt.Sprintf("Папка RK7: Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status)
			logger.Debug(c)
			err := CreateCateglistInWoo(categlistIndefiniteActive[i])
			if err != nil {
				errText = fmt.Sprintf("%s;%v", c, err)
				logger.Error(err.Error())
				resultSyncAll = append(resultSyncAll, errText)
				if errors.Unwrap(errors.Unwrap(err)).Error() == ERROR_CREATE_PRODUCTCATEGORY_EXIST {
					originalName := categlistIndefiniteWithWooIDActive[i].Name
					categlistIndefiniteActive[i].Name = fmt.Sprintf("%s_%d", categlistIndefiniteActive[i].Name, categlistIndefiniteActive[i].Code)
					text := fmt.Sprintf("Элемент с указанным именем уже существует у родительского элемента. Пробуем повторно создать с именем: %s", categlistIndefiniteActive[i].Name)
					logger.Warning(text)
					resultSyncAll = append(resultSyncAll, text)
					err := CreateCateglistInWoo(categlistIndefiniteActive[i])
					if err != nil {
						errText = fmt.Sprintf("%s;%v", c, err)
						logger.Error(err.Error())
						resultSyncAll = append(resultSyncAll, errText)
						resultSyncError = append(resultSyncAll, errText)
					} else {
						text := "Папка успешно создана"
						logger.Debug(text)
						resultSyncAll = append(resultSyncAll, text)
						flagNeedUpdate = true
					}
					categlistIndefiniteWithWooIDActive[i].Name = originalName
				} else {
					resultSyncError = append(resultSyncAll, errText)
				}
			} else {
				text := fmt.Sprintf("%s, папка успешно создана в WOO", c)
				logger.Debug(text)
				resultSyncAll = append(resultSyncAll, text)
				createCountInWoo++
			}
		}
		logger.Infof("Создано %d папок в WOO", createCountInWoo)
	}

	//++++
	if len(categlistIndefiniteWithWooIDNotActive) > 0 ||
		len(categlistIndefiniteNotActive) > 0 {
		logger.Info("Обнулить в RK7:")
		var nulledCountInRK7 int
		for i, categlist := range categlistIndefiniteWithWooIDNotActive {
			c := fmt.Sprintf("Папка RK7: Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status)
			logger.Debug(c)
			if categlist.WOO_ID == 0 && categlist.WOO_PARENT_ID == cfg.WOOCOMMERCE.MenuCategoryId {
				logger.Debugf("Обнуление WOO_ID/WOO_PARENT_ID в RK7 не требуется. WOO_ID=0, WOO_PARENT_ID=%d", cfg.WOOCOMMERCE.MenuCategoryId)
			} else {
				err = NulledCateglistInRK7(categlistIndefiniteWithWooIDNotActive[i])
				if err != nil {
					errText = fmt.Sprintf("%s;%v", c, err)
					logger.Error(err.Error())
					resultSyncAll = append(resultSyncAll, errText)
					resultSyncError = append(resultSyncAll, errText)
				} else {
					text := fmt.Sprintf("%s, папка успешно обнулена в RK7", c)
					logger.Debug(text)
					resultSyncAll = append(resultSyncAll, text)
					nulledCountInRK7++
				}
			}
		}
		for i, categlist := range categlistIndefiniteNotActive {
			c := fmt.Sprintf("Папка RK7: Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status)
			logger.Debug(c)
			if categlist.WOO_ID == 0 && categlist.WOO_PARENT_ID == cfg.WOOCOMMERCE.MenuCategoryId {
				logger.Debugf("Обнуление WOO_ID/WOO_PARENT_ID в RK7 не требуется. WOO_ID=0, WOO_PARENT_ID=%d", cfg.WOOCOMMERCE.MenuCategoryId)
			} else {
				err = NulledCateglistInRK7(categlistIndefiniteNotActive[i])
				if err != nil {
					errText = fmt.Sprintf("%s;%v", c, err)
					logger.Error(err.Error())
					resultSyncAll = append(resultSyncAll, errText)
					resultSyncError = append(resultSyncAll, errText)
				} else {
					logger.Debug("Папка успешно обнулена в RK7")
					text := fmt.Sprintf("%s, папка успешно обнулена в RK7", c)
					resultSyncAll = append(resultSyncAll, text)
					nulledCountInRK7++
				}
			}
		}
		logger.Infof("Обновлено %d папок в RK7", nulledCountInRK7)
	}

	if len(resultSyncError) > 0 {
		logger.Info("1-й этап синхронизации завершился с ошибками")
		if cfg.MENUSYNC.TelegramReport == 1 {
			telegram.SendMessageToTelegramWithLogError(strings.Join(resultSyncAll, "\n"))
		} else if cfg.MENUSYNC.TelegramReport == 2 {
			telegram.SendMessageToTelegramWithLogError(strings.Join(resultSyncError, "\n"))
		}
	} else {
		logger.Info("1-й этап синхронизации завершился успешно")
		logger.Info("Запущен 2-й этап синхронизации: свойства WOO_PARENT_ID/иерархия папок")
		// папки найденные в WOO
		var categlistNotActive []*modelsRK7API.Categlist                 // Папка не активна в RK7. Игнорируем
		var categlistNoNeedUpdateParentIdInWoo []*modelsRK7API.Categlist // Папка RK7 совпадает с WOO(свойства WOO_PARENT_ID). Обновление в WOO не требуется
		var categlistNeedUpdateParentIdInWoo []*modelsRK7API.Categlist   // Папка RK7 не совпадает с WOO(свойства WOO_PARENT_ID). Обновление в WOO требуется
		var categlistNotFoundCateglistParent []*modelsRK7API.Categlist   // Папка RK7 Parent: не найдена в кеше RK7

		// папки не найденные в WOO
		var categlistNotInCacheWithoutWooIDActive []*modelsRK7API.Categlist    // Папка RK7: не найдена в кеше WOO/активная. Папка должны быть с WOO_ID
		var categlistNotInCacheWithoutWooIDNotActive []*modelsRK7API.Categlist // Папка RK7: не найдена в кеше WOO/активная. Папка должны быть с WOO_ID
		var categlistNotWooIdActive []*modelsRK7API.Categlist                  // Папка RK7 без WOO_ID, активная. Папка должны быть с WOO_ID, сообщаем об ошибке
		var categlistNotWooIdNotActive []*modelsRK7API.Categlist               // Папка RK7 без WOO_ID, не активная. Cообщаем без ошибок// TODO resultSyncAll = append(resultSyncAll, errText)

		// необходимо обнcateglistNoNeedUpdateNameInWooовить наименование папок
		var categlistNoNeedUpdateNameInWoo []*modelsRK7API.Categlist
		var categlistNeedUpdateNameInWoo []*modelsRK7API.Categlist

	LoopTwoStage:
		for i, categlist := range categlists {
			logger.Debugf("Папка RK7: Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status)

			for _, ignoreIdent := range cfg.RK7.CateglistIdentIgnore {
				if categlist.Ident == ignoreIdent {
					logger.Debug("Игнорируем по настройкам конфига")
					logger.Debug("--------------------------------------")
					continue LoopTwoStage
				}
			}

			if categlist.WOO_ID != 0 {
				logger.Debug("Папка с не пустым WOO_ID. Соответственно имеется в WOO в рамкам проверки 1-го этапа синхронизации папок")
				if category, found := categoriesWooByID[categlist.WOO_ID]; found {
					logger.Debug("Папка найдена в кеше WOO")
					if categlist.Status != 3 {
						logger.Debug("Папка не активна в RK7. Игнорируем")
						categlistNotActive = append(categlistNotActive, categlists[i])
					} else {
						logger.Debug("Папка активна в RK7. Необходимо сравнить с WOO(свойство WOO_PARENT_ID)")

						if categlistParent, found := categlistsRK7ByIdent[categlist.MainParentIdent]; found {
							logger.Debug("Папка Parent найдена в кеше RK7")
							var parentID int
							if categlistParent.ItemIdent == 0 {
								logger.Debug("Папка Parent корневая - используем WOO_ID из cfg.WOOCOMMERCE.cfg.WOOCOMMERCE.MenuCategoryId")
								parentID = cfg.WOOCOMMERCE.MenuCategoryId
							} else {
								logger.Debug("Папка Parent не корневая - используем WOO_ID из categlistsRK7ByIdent[categlist.MainParentIdent]")
								parentID = categlistParent.WOO_ID
							}

							logger.Debugf("RK.WOO_PARENT_ID=%d && WOO.Parent=%d", categlist.WOO_PARENT_ID, category.Parent)
							logger.Debugf("RK.WOO_PARENT_ID=%d && RK.MainParent.WOO_ID=%d", categlist.WOO_PARENT_ID, parentID)
							if categlist.WOO_PARENT_ID == category.Parent && categlist.WOO_PARENT_ID == parentID {
								logger.Debug("Папка RK7 совпадает с WOO(свойства WOO_PARENT_ID). Обновление в WOO не требуется")
								categlistNoNeedUpdateParentIdInWoo = append(categlistNoNeedUpdateParentIdInWoo, categlists[i])
							} else {
								logger.Debug("Папка RK7 не совпадает с WOO(свойства WOO_PARENT_ID). Обновление в WOO требуется")
								categlistNeedUpdateParentIdInWoo = append(categlistNeedUpdateParentIdInWoo, categlists[i])
							}

							logger.Debugf("RK.Name=%s && WOO.Name=%s", categlist.Name, category.Name)
							if categlist.Name == category.Name {
								logger.Debug("Папка RK7 совпадает с WOO(свойство Name). Обновление в WOO не требуется")
								categlistNoNeedUpdateNameInWoo = append(categlistNoNeedUpdateNameInWoo, categlists[i])
							} else {
								logger.Debug("Папка RK7 не совпадает с WOO(свойства Name). Обновление в WOO требуется")
								categlistNeedUpdateNameInWoo = append(categlistNeedUpdateNameInWoo, categlists[i])
							}

						} else {
							errText = fmt.Sprintf("Папка RK7 Parent не найдена в кеше RK7: Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status)
							logger.Error(err.Error())
							resultSyncAll = append(resultSyncAll, errText)
							categlistNotFoundCateglistParent = append(categlistNotFoundCateglistParent, categlists[i])
						}
					}
				} else {
					if categlist.Status == 3 {
						errText = fmt.Sprintf("Папка RK7: не найдена в кеше WOO/активная. Папка должны быть с WOO_ID: Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status)
						categlistNotInCacheWithoutWooIDActive = append(categlistNotInCacheWithoutWooIDActive, categlists[i])
					} else {
						errText = fmt.Sprintf("Папка RK7: не найдена в кеше WOO/активная. Папка должны быть с WOO_ID: Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status)
						categlistNotInCacheWithoutWooIDNotActive = append(categlistNotInCacheWithoutWooIDNotActive, categlists[i])
					}
					logger.Error(err.Error())
					resultSyncAll = append(resultSyncAll, errText)
				}
			} else {
				if categlist.Status == 3 {
					errText = fmt.Sprintf("Папка RK7 без WOO_ID, активная. Папка должны быть с WOO_ID: Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status)
					categlistNotWooIdActive = append(categlistNotWooIdActive, categlists[i])
					logger.Error(err.Error())
					resultSyncError = append(resultSyncAll, errText)
					resultSyncAll = append(resultSyncAll, errText)
				} else {
					errText = fmt.Sprintf("Папка RK7 без WOO_ID, не активная. Папка должны быть с WOO_ID: Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status)
					categlistNotWooIdNotActive = append(categlistNotWooIdNotActive, categlists[i])
					logger.Warning(errText)
					resultSyncAll = append(resultSyncAll, errText)
				}
			}
			logger.Debug("--------------------------------------")
		}

		logger.Info("Синхронизация иерархии папок WOO:")
		logger.Infof("Всего: %d", len(categlists))
		logger.Infof("Не активные: %d", len(categlistNotActive))

		logger.Infof("Найдено в WOO: %d", len(categlistNoNeedUpdateParentIdInWoo)+len(categlistNeedUpdateParentIdInWoo)+len(categlistNotFoundCateglistParent))
		logger.Infof("Совпадают с WOO: %d", len(categlistNoNeedUpdateParentIdInWoo))
		logger.Infof("Обновить в WOO: %d", len(categlistNeedUpdateParentIdInWoo))
		logger.Infof("Папка RK7 Parent не найдена в кеше RK7(ошибка): %d", len(categlistNotFoundCateglistParent))

		logger.Infof("Не найдено в WOO - некорректная ситуация(ошибка): %d", len(categlistNotInCacheWithoutWooIDActive)+len(categlistNotInCacheWithoutWooIDNotActive)+len(categlistNotWooIdActive)+len(categlistNotWooIdNotActive))
		logger.Infof("WOO_ID неопределен, активная(ошибка): %d", len(categlistNotInCacheWithoutWooIDActive))
		logger.Infof("WOO_ID неопределен, не активная(ошибка): %d", len(categlistNotInCacheWithoutWooIDNotActive))
		logger.Infof("Без WOO_ID, активная(ошибка): %d", len(categlistNotWooIdActive))
		logger.Infof("Без WOO_ID, не активная(ошибка): %d", len(categlistNotWooIdNotActive))

		logger.Info("Папки WOO:")
		logger.Infof("ProductCategoriesWooByID: %d", len(categoriesWooByID))
		logger.Infof("ProductCategoriesWooBySlug: %d", len(categoriesWooBySlug))

		logger.Infof("Необходимо обновить Name: %d", len(categlistNeedUpdateNameInWoo))
		logger.Infof("Обновить Name не требуется: %d", len(categlistNoNeedUpdateNameInWoo))

		logger.Info("Обновляем в WOO и RK7:")
		var updateCountInWoo, updateCountInRK7 int
		for i, categlist := range categlistNeedUpdateParentIdInWoo {
			c := fmt.Sprintf("Папка RK7: Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status)
			logger.Debug(c)
			if categlistParent, found := categlistsRK7ByIdent[categlist.MainParentIdent]; found {
				logger.Debug("Папка Parent найдена в кеше RK7")
				var parentID int
				if categlistParent.ItemIdent == 0 {
					logger.Debug("Папка Parent корневая - используем WOO_ID из cfg.WOOCOMMERCE.cfg.WOOCOMMERCE.MenuCategoryId")
					parentID = cfg.WOOCOMMERCE.MenuCategoryId
				} else {
					logger.Debug("Папка Parent не корневая - используем WOO_ID из categlistsRK7ByIdent[categlist.MainParentIdent]")
					parentID = categlistParent.WOO_ID
				}

				logger.Debugf("RK.WOO_PARENT_ID=%d && RK.MainParent.WOO_ID=%d", categlist.WOO_PARENT_ID, parentID)
				if categlist.WOO_PARENT_ID == parentID {
					logger.Debug("WOO_PARENT_ID актуальный. Обновление в RK7 не требуется")
				} else {
					logger.Debug("WOO_PARENT_ID не актуальный. Обновление в RK7 требуется")
					err = UpdateParentIDCateglistInRK7(categlistNeedUpdateParentIdInWoo[i], parentID)
					if err != nil {
						errText = fmt.Sprintf("%s;%v", c, err)
						logger.Error(err.Error())
						resultSyncAll = append(resultSyncAll, errText)
						resultSyncError = append(resultSyncAll, errText)
					} else {
						logger.Debug("Папка успешно обновлена в RK7")
						text := fmt.Sprintf("%s, папка успешно обновлена в RK7", c)
						resultSyncAll = append(resultSyncAll, text)
						updateCountInRK7++
						logger.Debug("Выполняем поиск папки в WOO")
						if _, found := categoriesWooByID[categlist.WOO_ID]; found {
							logger.Debug("Папка найдена в WOO. Выполняем обновление")
							err = UpdateCateglistInWoo(categlistNeedUpdateParentIdInWoo[i])
							if err != nil {
								errText = fmt.Sprintf("%s;%v", c, err)
								logger.Error(err.Error())
								resultSyncAll = append(resultSyncAll, errText)
								resultSyncError = append(resultSyncAll, errText)
							} else {
								logger.Debug("Папка успешно обновлена в WOO")
								text := fmt.Sprintf("%s, папка успешно обновлена в WOO", c)
								resultSyncAll = append(resultSyncAll, text)
								updateCountInWoo++
							}
						} else {
							logger.Error("Папка не найдена в кеше WOO. Обновление не выполнить")
							errText = fmt.Sprintf("%s;%v", c, "Папка не найдена в кеше WOO. Обновление не выполнить")
							resultSyncAll = append(resultSyncAll, errText)
							resultSyncError = append(resultSyncAll, errText)
						}
					}
				}
			} else {
				errText = "Папка RK7 Parent: не найдена в кеше RK7"
				logger.Error(err.Error())
				telegramText := fmt.Sprintf("%s;%v", c, errText)
				resultSyncAll = append(resultSyncAll, telegramText)
				resultSyncError = append(resultSyncAll, telegramText)
			}
		}
		logger.Infof("Обновлено %d папок в WOO", updateCountInWoo)
		logger.Infof("Обновлено %d папок в RK7", updateCountInRK7)
		if updateCountInWoo != updateCountInRK7 {
			errText = fmt.Sprintf("Не совпадает количество папок обновления в WOO=%d и RK7=%d", updateCountInWoo, updateCountInRK7)
			logger.Error(err.Error())
			telegramText := errText
			resultSyncAll = append(resultSyncAll, telegramText)
			resultSyncError = append(resultSyncAll, telegramText)
		}

		//++
		logger.Info("Обновляем Name в WOO и RK7:")
		for i, categlist := range categlistNeedUpdateNameInWoo {
			c := fmt.Sprintf("Папка RK7: Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status)
			logger.Debug(c)
			logger.Debug("Выполняем поиск папки в WOO")
			if _, found := categoriesWooByID[categlist.WOO_ID]; found {
				err := UpdateCateglistInWoo(categlistNeedUpdateNameInWoo[i])
				if err != nil {
					errText = fmt.Sprintf("%s;%v", c, err)
					logger.Error(err.Error())
					resultSyncAll = append(resultSyncAll, errText)
					resultSyncError = append(resultSyncAll, errText)
				} else {
					logger.Debug("Папка успешно обновлена в WOO")
					text := fmt.Sprintf("%s, папка успешно обновлена в WOO", c)
					resultSyncAll = append(resultSyncAll, text)
					updateCountInWoo++
				}
			} else {
				logger.Error("Папка не найдена в кеше WOO. Обновление не выполнить")
				errText = fmt.Sprintf("%s;%v", c, "Папка не найдена в кеше WOO. Обновление не выполнить")
				resultSyncAll = append(resultSyncAll, errText)
				resultSyncError = append(resultSyncAll, errText)
			}
		}

		if len(resultSyncError) > 0 {
			logger.Info("2-й этап синхронизации завершился с ошибками")
			if cfg.MENUSYNC.TelegramReport == 1 {
				telegram.SendMessageToTelegramWithLogError(strings.Join(resultSyncAll, "\n"))
			} else if cfg.MENUSYNC.TelegramReport == 2 {
				telegram.SendMessageToTelegramWithLogError(strings.Join(resultSyncError, "\n"))
			}
		} else {
			logger.Info("2-й этап синхронизации завершился успешно")
			if cfg.MENUSYNC.TelegramReport == 1 {
				telegram.SendMessageToTelegramWithLogError("2-й этап синхронизации завершился успешно")
			}
			VersionRefName, err := GetVersion(rk7API, "Categlist")
			if err != nil {
				return errors.Wrapf(err, "failed GetVersion(RK7API, %s)", "Categlist")
			}

			err = UpdateVersionInDB(DB, "Categlist", VersionRefName)
			if err != nil {
				return errors.Wrapf(err, "failed UpdateVersionInDB(DB, %s, %d)", "Categlist", VersionRefName)
			}

			logger.Info("Результат обновление папок:")
			logger.Infof("Длина categlists(все): %d", len(categlists))
			logger.Infof("Длина categlistsRK7ByIdent(все): %d", len(categlistsRK7ByIdent))
			logger.Infof("Длина categoriesWooByID: %d", len(categoriesWooByID))
		}
	}

	if flagNeedUpdate {
		return errors.New(SYNC_CATEGLIST_NEED_UPDATE)
	} else {
		return nil
	}
}

const SYNC_CATEGLIST_NEED_UPDATE = "Need update"

//++
// удалить папку в Woo
func DeleteCateglistInWoo(categlist *modelsRK7API.Categlist) error {
	//TODO при удалении может быть не найдено блюдо, нужно будет тогда проигнорировать ошибку

	logger := logging.GetLogger()
	logger.Debug("Start DeleteCateglistInWoo")
	defer logger.Debug("End DeleteCateglistInWoo")

	var err error
	woo := wooapi.GetAPI()
	menu, err := cache.GetMenu()
	if err != nil {
		return errors.Wrap(err, "Ошибка при получении кеша меню")
	}
	logger.Debugf("Удаляем папку из WOO/кеша WOO")
	err = woo.ProductCategoryDelete(categlist.WOO_ID)
	if err != nil {
		return errors.Wrap(err, "Ошибка при удалении папки из WOO")
	} else {
		logger.Debug("Обнуляем кеш WOO")
		err = menu.DeleteProductCategoryFromCache(categlist.WOO_ID)
		if err != nil {
			return errors.Wrap(err, "Ошибка при удалении папки из кеша WOO")
		} else {
			logger.Debug("Обнулен кеш WOO. Папка успешно удалена из WOO.")
			err = NulledCateglistInRK7(categlist)
			if err != nil {
				return errors.Wrap(err, "failed NulledCateglistInRK7(categlist)")
			} else {
				return nil
			}
		}
	}
}

//++
// обновить папку в Woo - свойство Name и Parent
func UpdateCateglistInWoo(categlist *modelsRK7API.Categlist) error {
	//TODO если при обновлении не найдено блюдо, то необходимо его создать и после обновить папку RK7

	logger := logging.GetLogger()
	logger.Debug("Start UpdateCateglistInWoo")
	defer logger.Debug("End UpdateCateglistInWoo")

	var err error
	woo := wooapi.GetAPI()
	if err != nil {
		return errors.Wrap(err, "Ошибка при получении кеша меню")
	}

	logger.Debugf("Обновляем папку в WOO/кеше WOO")
	category, err := woo.ProductCategoryGet(categlist.WOO_ID)
	if err != nil {
		return errors.Wrapf(err, "Ошибка при получении ProductCategoryGet(ID=%d)", categlist.WOO_ID)
	} else {
		if category != nil {
			logger.Debugf("ProductCategory успешно получен: Name=%s, ID=%d, Parent=%d, Slug=%s", category.Name, category.ID, category.Parent, category.Slug)

			var categlistName string
			if categlist.WOO_LONGNAME != "" {
				categlistName = categlist.WOO_LONGNAME
			} else {
				categlistName = categlist.Name
			}

			recoveryName := category.Name
			if category.Name != categlistName {
				category.Name = categlistName
			}
			recoveryParent := category.Parent
			if category.Parent != categlist.WOO_PARENT_ID {
				category.Parent = categlist.WOO_PARENT_ID
			}

			_, err = woo.ProductCategoryUpdate(category)
			if err != nil {
				category.Name = recoveryName
				category.Parent = recoveryParent
				return errors.Wrap(err, "Ошибка при обновлении папки. Кеш восстановлен")
			} else {
				logger.Debug("Папка успешно обновлена. Кеш обновлен")
				return nil
			}
		} else {
			return errors.New(fmt.Sprintf("Не удалось получить ProductCategoryGet(ID=%d)", categlist.WOO_ID))
		}
	}
}

const ERROR_CREATE_PRODUCTCATEGORY_EXIST = "code:term_exists; message:Элемент с указанным именем уже существует у родительского элемента.; status:400; display:; details:;"

//++
// создать папку в Woo
func CreateCateglistInWoo(categlist *modelsRK7API.Categlist) error {
	//TODO подумать какие ошибки могут быть например если не удастся создать в WOO потому что существует

	logger := logging.GetLogger()
	logger.Debug("Start CreateCateglistInWoo")
	defer logger.Debug("End CreateCateglistInWoo")

	var err error
	woo := wooapi.GetAPI()
	if err != nil {
		return errors.Wrap(err, "Ошибка при получении кеша меню")
	}
	rk7 := rk7api.GetAPI("REF")
	menu, err := cache.GetMenu()
	if err != nil {
		return errors.Wrap(err, "Ошибка при получении кеша меню")
	}
	cfg := config.GetConfig()

	logger.Infof("Создаем папку в WOO/кеше WOO")
	category := new(modelsWOOAPI.ProductCategory)
	var categlistName string
	if categlist.WOO_LONGNAME != "" {
		categlistName = categlist.WOO_LONGNAME
	} else {
		categlistName = categlist.Name
	}
	category.Name = categlistName
	category.Parent = cfg.WOOCOMMERCE.MenuCategoryId
	categoryCreated, err := woo.ProductCategoryAdd(category)
	if err != nil {
		return errors.Wrapf(err, "Ошибка при создании папки в WOO; ProductCategoryAdd(Name=%s)", categlist.Name)
	} else {
		if categoryCreated != nil {
			logger.Debugf("Папка в WOO создана успешно: Name=%s, ID=%d, Parent=%d, Slug=%s",
				categoryCreated.Name,
				categoryCreated.ID,
				categoryCreated.Parent,
				categoryCreated.Slug)
			logger.Debug("Обновляем кеш WOO")
			err = menu.AddProductCategoryToCache(categoryCreated)
			if err != nil {
				return errors.Wrap(err, "Ошибка при добавление папки в кеш WOO")
			} else {
				logger.Debug("Обновлен кеш WOO. Обновляем свойства в RK7")
				var categlists []*modelsRK7API.Categlist
				categlist.WOO_ID = categoryCreated.ID
				categlist.WOO_PARENT_ID = cfg.WOOCOMMERCE.MenuCategoryId
				categlists = append(categlists, categlist)
				_, err = rk7.SetRefDataCateglist(categlists)
				if err != nil {
					categlist.WOO_ID = 0
					categlist.WOO_PARENT_ID = cfg.WOOCOMMERCE.MenuCategoryId
					return errors.Wrap(err, "Ошибка при обновлении WOO_ID/WOO_PARENT_ID в RK7. Кеш установлен по умолчанию.")
				} else {
					logger.Debug("Папка успешно обновлена")
					return nil
				}
			}
		} else {
			return errors.New(fmt.Sprintf("Не удалось создать папку в WOO; ProductCategoryAdd(Name=%d)", categlist.Name))
		}
	}
}

//++
// обнулить папку в RK7 - свойства WOO_ID и WOO_PARENT_ID
func NulledCateglistInRK7(categlist *modelsRK7API.Categlist) error {
	logger := logging.GetLogger()
	logger.Debug("Start NulledCateglistInRK7")
	defer logger.Debug("End NulledCateglistInRK7")
	var err error
	rk7 := rk7api.GetAPI("REF")
	cfg := config.GetConfig()
	logger.Debug("Обнуляем WOO_ID/WOO_PARENT_ID в RK7.")

	var categlists []*modelsRK7API.Categlist
	recoveryWooID := categlist.WOO_ID
	recoveryWooParentID := categlist.WOO_PARENT_ID
	categlist.WOO_ID = 0
	categlist.WOO_PARENT_ID = cfg.WOOCOMMERCE.MenuCategoryId
	categlists = append(categlists, categlist)
	_, err = rk7.SetRefDataCateglist(categlists)
	if err != nil {
		categlist.WOO_ID = recoveryWooID
		categlist.WOO_PARENT_ID = recoveryWooParentID
		return errors.Wrap(err, "Ошибка при обнулении WOO_ID/WOO_PARENT_ID в RK7. Кеш восстановлен")
	} else {
		logger.Debug("Папка успешно обновлена")
		return nil
	}
}

//++
// обновить папку в RK7 - свойство WOO_PARENT_ID
func UpdateParentIDCateglistInRK7(categlist *modelsRK7API.Categlist, parentID int) error {
	logger := logging.GetLogger()
	logger.Debug("Start UpdateParentIDCateglistInRK7")
	defer logger.Debug("End UpdateParentIDCateglistInRK7")
	var err error
	rk7 := rk7api.GetAPI("REF")
	logger.Debug("Обновляем WOO_PARENT_ID в RK7/кеше RK7")

	var categlists []*modelsRK7API.Categlist
	recoveryWooParentID := categlist.WOO_PARENT_ID
	categlist.WOO_PARENT_ID = parentID
	categlists = append(categlists, categlist)
	_, err = rk7.SetRefDataCateglist(categlists)
	if err != nil {
		categlist.WOO_PARENT_ID = recoveryWooParentID
		return errors.Wrap(err, "Ошибка при обновлении WOO_PARENT_ID в RK7. Кеш восстановлен")
	} else {
		logger.Debug("Папка успешно обновлена в RK7/кеше RK7")
		return nil
	}
}
