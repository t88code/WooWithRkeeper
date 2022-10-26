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

	err = menu.RefreshCateglist()
	if err != nil {
		return err
	}

	err = menu.RefreshProductCategories()
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

		for _, ignoreIdent := range cfg.RK7MID.CateglistIdentIgnore {
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
	logger.Infof("Игнорировано: %d", len(cfg.RK7MID.CateglistIdentIgnore))

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

	logger.Info("Удаляем в WOO:")
	var delCountInWoo int
	for i, categlist := range categlistNeedDelInWoo {
		logger.Debugf("Папка RK7: Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status)
		err = DeleteCateglistInWoo(categlistNeedDelInWoo[i])
		if err != nil {
			logger.Error(err.Error())
			errSync = append(errSync, err.Error())
		} else {
			logger.Info("Папка успешно удалена в WOO")
			delCountInWoo++
		}
	}
	logger.Infof("Удалено %d папок в WOO", delCountInWoo)

	logger.Info("Обновляем в WOO:")
	var updateCountInWoo int
	for i, categlist := range categlistNeedUpdateInWoo {
		logger.Debugf("Папка RK7: Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status)
		err = UpdateCateglistInWoo(categlistNeedUpdateInWoo[i])
		if err != nil {
			logger.Error(err.Error())
			errSync = append(errSync, err.Error())
		} else {
			logger.Info("Папка успешно обновлена в WOO")
			updateCountInWoo++
		}
	}
	logger.Infof("Обновлено %d папок в WOO", updateCountInWoo)

	logger.Info("Создаем в WOO:")
	var createCountInWoo int
	for i, categlist := range categlistIndefiniteWithWooIDActive {
		logger.Debugf("Папка RK7: Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status)
		err = CreateCateglistInWoo(categlistIndefiniteWithWooIDActive[i])
		if err != nil {
			logger.Error(err.Error())
			errSync = append(errSync, err.Error())
		} else {
			logger.Info("Папка успешно создана в WOO")
			createCountInWoo++
		}
	}
	for i, categlist := range categlistIndefiniteActive {
		logger.Debugf("Папка RK7: Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status)
		err = CreateCateglistInWoo(categlistIndefiniteActive[i])
		if err != nil {
			logger.Error(err.Error())
			errSync = append(errSync, err.Error())
		} else {
			logger.Info("Папка успешно создана в WOO")
			createCountInWoo++
		}
	}
	logger.Infof("Создано %d папок в WOO", createCountInWoo)

	logger.Info("Обнулить в RK7:")
	var nulledCountInRK7 int
	for i, categlist := range categlistIndefiniteWithWooIDNotActive {
		logger.Debugf("Папка RK7: Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status)
		if categlist.WOO_ID == 0 && categlist.WOO_PARENT_ID == cfg.WOOCOMMERCE.MenuCategoryId {
			logger.Infof("Обнуление WOO_ID/WOO_PARENT_ID в RK7 не требуется. WOO_ID=0, WOO_PARENT_ID=%d", cfg.WOOCOMMERCE.MenuCategoryId)
		} else {
			err = NulledCateglistInRK7(categlistIndefiniteWithWooIDNotActive[i])
			if err != nil {
				logger.Error(err.Error())
				errSync = append(errSync, err.Error())
			} else {
				logger.Info("Папка успешно обнулена в RK7")
				nulledCountInRK7++
			}
		}

	}
	for i, categlist := range categlistIndefiniteNotActive {
		logger.Debugf("Папка RK7: Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status)
		if categlist.WOO_ID == 0 && categlist.WOO_PARENT_ID == cfg.WOOCOMMERCE.MenuCategoryId {
			logger.Infof("Обнуление WOO_ID/WOO_PARENT_ID в RK7 не требуется. WOO_ID=0, WOO_PARENT_ID=%d", cfg.WOOCOMMERCE.MenuCategoryId)
		} else {
			err = NulledCateglistInRK7(categlistIndefiniteNotActive[i])
			if err != nil {
				logger.Error(err.Error())
				errSync = append(errSync, err.Error())
			} else {
				logger.Info("Папка успешно обнулена в RK7")
				nulledCountInRK7++
			}
		}
	}
	logger.Infof("Обновлено %d папок в RK7", nulledCountInRK7)

	if len(errSync) > 0 && cfg.MENUSYNC.TelegramReport == 1 {
		logger.Info("1-й этап синхронизации завершился с ошибками")
		telegram.SendMessageToTelegramWithLogError(strings.Join(errSync, "\n"))
	} else {
		logger.Info("1-й этап синхронизации завершился успешно")
		logger.Info("Запущен 2-й этап синхронизации: свойства WOO_PARENT_ID/иерархия папок")

		// todo нужно ли обновлять кеш или не стоит - подумать

		// папки найденные в WOO
		var categlistNotActive []*modelsRK7API.Categlist                 // Папка не активна в RK7. Игнорируем
		var categlistNoNeedUpdateParentIdInWoo []*modelsRK7API.Categlist // Папка RK7 совпадает с WOO(свойства WOO_PARENT_ID). Обновление в WOO не требуется
		var categlistNeedUpdateParentIdInWoo []*modelsRK7API.Categlist   // Папка RK7 не совпадает с WOO(свойства WOO_PARENT_ID). Обновление в WOO требуется
		var categlistNotFoundCateglistParent []*modelsRK7API.Categlist   // Папка RK7 Parent: не найдена в кеше RK7

		// папки не найденные в WOO
		var categlistNotInCacheWithoutWooIDActive []*modelsRK7API.Categlist    // Папка RK7: не найдена в кеше WOO/активная. Папка должны быть с WOO_ID
		var categlistNotInCacheWithoutWooIDNotActive []*modelsRK7API.Categlist // Папка RK7: не найдена в кеше WOO/активная. Папка должны быть с WOO_ID
		var categlistNotWooIdActive []*modelsRK7API.Categlist                  // Папка RK7 без WOO_ID, активная. Папка должны быть с WOO_ID, сообщаем об ошибке
		var categlistNotWooIdNotActive []*modelsRK7API.Categlist               // Папка RK7 без WOO_ID, не активная. Папка должны быть с WOO_ID, сообщаем об ошибке

	LoopTwoStage:
		for i, categlist := range categlists {
			logger.Debugf("Папка RK7: Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status)

			for _, ignoreIdent := range cfg.RK7MID.CateglistIdentIgnore {
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
						} else {
							errText = fmt.Sprintf("Папка RK7 Parent не найдена в кеше RK7: Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status)
							logger.Error(errText)
							errSync = append(errSync, errText)
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
					logger.Error(errText)
					errSync = append(errSync, errText)
				}
			} else {
				if categlist.Status == 3 {
					errText = fmt.Sprintf("Папка RK7 без WOO_ID, активная. Папка должны быть с WOO_ID: Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status)
					categlistNotWooIdActive = append(categlistNotWooIdActive, categlists[i])
				} else {
					errText = fmt.Sprintf("Папка RK7 без WOO_ID, не активная. Папка должны быть с WOO_ID: Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status)
					categlistNotWooIdNotActive = append(categlistNotWooIdNotActive, categlists[i])
				}
				logger.Error(errText)
				errSync = append(errSync, errText)
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

		logger.Info("Обновляем в WOO и RK7:")
		var updateCountInWoo, updateCountInRK7 int
		for i, categlist := range categlistNeedUpdateParentIdInWoo {
			logger.Debugf("Папка RK7: Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status)

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
					err = UpdateCateglistInRK7(categlistNeedUpdateParentIdInWoo[i], parentID)
					if err != nil {
						logger.Error(err.Error())
						errSync = append(errSync, err.Error())
					} else {
						logger.Info("Папка успешно обновлена в RK7")
						updateCountInRK7++
						logger.Info("Выполняем поиск папки в WOO")
						if _, found := categoriesWooByID[categlist.WOO_ID]; found {
							logger.Debug("Папка найдена в WOO. Выполняем обновление")
							err = UpdateCateglistInWoo(categlistNeedUpdateParentIdInWoo[i])
							if err != nil {
								logger.Error(err.Error())
								errSync = append(errSync, err.Error())
							} else {
								logger.Info("Папка успешно обновлена в WOO")
								updateCountInWoo++
							}
						} else {
							errText = "Папка не найдена в кеше WOO. Обновление не выполнить"
							logger.Error(errText)
							errSync = append(errSync, errText)
						}
					}
				}
			} else {
				errText = "Папка RK7 Parent: не найдена в кеше RK7"
				logger.Error(errText)
				errSync = append(errSync, errText)
			}
		}
		logger.Infof("Обновлено %d папок в WOO", updateCountInWoo)
		logger.Infof("Обновлено %d папок в RK7", updateCountInRK7)
		if updateCountInWoo != updateCountInRK7 {
			errText = fmt.Sprintf("Не совпадает количество папок обновления в WOO=%d и RK7=%d", updateCountInWoo, updateCountInRK7)
			logger.Error(errText)
			errSync = append(errSync, errText)
		}

		if len(errSync) > 0 && cfg.MENUSYNC.TelegramReport == 1 {
			logger.Info("2-й этап синхронизации завершился с ошибками")
			telegram.SendMessageToTelegramWithLogError(strings.Join(errSync, "\n"))
		} else {
			logger.Info("2-й этап синхронизации завершился успешно")

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

	return nil
}

// удалить папку в Woo
func DeleteCateglistInWoo(categlist *modelsRK7API.Categlist) error {
	//TODO при удалении может быть не найдено блюдо, нужно будет тогда проигнорировать ошибку

	logger := logging.GetLogger()
	logger.Info("Start DeleteCateglistInWoo")
	defer logger.Info("End DeleteCateglistInWoo")

	var err error
	woo := wooapi.GetAPI()
	menu, err := cache.GetMenu()
	if err != nil {
		return errors.Wrap(err, "Ошибка при получении кеша меню")
	}
	logger.Infof("Удаляем папку из WOO/кеша WOO")
	err = woo.ProductCategoryDelete(categlist.WOO_ID)
	if err != nil {
		return errors.Wrap(err, "Ошибка при удалении папки из WOO")
	} else {
		logger.Info("Обнуляем кеш WOO")
		err = menu.DeleteProductCategoryFromCache(categlist.WOO_ID)
		if err != nil {
			return errors.Wrap(err, "Ошибка при удалении папки из кеша WOO")
		} else {
			logger.Info("Обнулен кеш WOO. Папка успешно удалена из WOO.")
			err = NulledCateglistInRK7(categlist)
			if err != nil {
				return errors.Wrap(err, "failed NulledCateglistInRK7(categlist)")
			} else {
				return nil
			}
		}
	}
}

// обновить папку в Woo - свойство Name и Parent
func UpdateCateglistInWoo(categlist *modelsRK7API.Categlist) error {
	//TODO если при обновлении не найдено блюдо, то необходимо его создать и после обновить папку RK7

	logger := logging.GetLogger()
	logger.Info("Start UpdateCateglistInWoo")
	defer logger.Info("End UpdateCateglistInWoo")

	var err error
	woo := wooapi.GetAPI()
	if err != nil {
		return errors.Wrap(err, "Ошибка при получении кеша меню")
	}

	logger.Infof("Обновляем папку в WOO/кеше WOO")
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
				logger.Info("Папка успешно обновлена. Кеш обновлен")
				return nil
			}
		} else {
			return errors.New(fmt.Sprintf("Не удалось получить ProductCategoryGet(ID=%d)", categlist.WOO_ID))
		}
	}
}

// создать папку в Woo
func CreateCateglistInWoo(categlist *modelsRK7API.Categlist) error {
	//TODO подумать какие ошибки могут быть например если не удастся создать в WOO потому что существует

	logger := logging.GetLogger()
	logger.Info("Start CreateCateglistInWoo")
	defer logger.Info("End CreateCateglistInWoo")

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
			logger.Info("Обновляем кеш WOO")
			err = menu.AddProductCategoryToCache(categoryCreated)
			if err != nil {
				return errors.Wrap(err, "Ошибка при добавление папки в кеш WOO")
			} else {
				logger.Info("Обновлен кеш WOO. Обновляем свойства в RK7")
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
					logger.Info("Папка успешно обновлена")
					return nil
				}
			}
		} else {
			return errors.New(fmt.Sprintf("Не удалось создать папку в WOO; ProductCategoryAdd(Name=%d)", categlist.Name))
		}
	}
}

// обнулить папку в RK7 - свойства WOO_ID и WOO_PARENT_ID
func NulledCateglistInRK7(categlist *modelsRK7API.Categlist) error {
	logger := logging.GetLogger()
	logger.Info("Start NulledCateglistInRK7")
	defer logger.Info("End NulledCateglistInRK7")
	var err error
	rk7 := rk7api.GetAPI()
	cfg := config.GetConfig()
	logger.Info("Обнуляем WOO_ID/WOO_PARENT_ID в RK7.")

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
		logger.Info("Папка успешно обновлена")
		return nil

	}
}

// обновить папку в RK7 - свойство WOO_PARENT_ID
func UpdateCateglistInRK7(categlist *modelsRK7API.Categlist, parentID int) error {
	logger := logging.GetLogger()
	logger.Info("Start UpdateCateglistInRK7")
	defer logger.Info("End UpdateCateglistInRK7")
	var err error
	rk7 := rk7api.GetAPI()
	logger.Info("Обновляем WOO_PARENT_ID в RK7/кеше RK7")

	var categlists []*modelsRK7API.Categlist
	recoveryWooParentID := categlist.WOO_PARENT_ID
	categlist.WOO_PARENT_ID = parentID
	categlists = append(categlists, categlist)
	_, err = rk7.SetRefDataCateglist(categlists)
	if err != nil {
		categlist.WOO_PARENT_ID = recoveryWooParentID
		return errors.Wrap(err, "Ошибка при обновлении WOO_PARENT_ID в RK7. Кеш восстановлен")
	} else {
		logger.Info("Папка успешно обновлена в RK7/кеше RK7")
		return nil
	}
}
