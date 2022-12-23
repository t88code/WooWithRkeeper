package categlist

import (
	"WooWithRkeeper/internal/cache"
	"WooWithRkeeper/internal/config"
	categlistDatabase "WooWithRkeeper/internal/database/model/categlist"
	"WooWithRkeeper/internal/rk7api"
	"WooWithRkeeper/internal/rk7api/models"
	"WooWithRkeeper/internal/telegram"
	"WooWithRkeeper/internal/wooapi"
	modelsWOOAPI "WooWithRkeeper/internal/wooapi/models"
	"WooWithRkeeper/internal/wooapi/options"
	"WooWithRkeeper/pkg/logging"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"strings"
)

type handler struct {
	status  string
	handler func(*sqlx.DB, string, string) error
	message string
}

func (h *handler) Do(db *sqlx.DB) error {
	logger := logging.GetLogger()
	logger.Debugf("Start doHandler.%s", h.status)
	defer logger.Debugf("Start doHandler.%s", h.status)

	return h.handler(db, h.status, h.message)
}

// GetCateglistDescription Сформировать строку с блюдом
// Используется в 1 этапе синхронизации папок
func GetCateglistDescription(categlist *models.Categlist) string {
	return fmt.Sprintf("Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d, WOO_SYNC %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status, categlist.WOO_SYNC)
}

func GetProductCategoryDescription(productCategory *modelsWOOAPI.ProductCategory) string {
	return fmt.Sprintf("Name: %s, ID: %d, Parent: %d", productCategory.Name, productCategory.ID, productCategory.Parent)
}

// HandlerCateglistToDb 1 этап - закачка RK7.Categlist/WOO.ProductCategory в DB.Categlist
// todo messageError to telegram
func HandlerCateglistToDb(db *sqlx.DB) error {
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

		categlistRowDb := categlistDatabase.Categlist{
			Name: sql.NullString{
				String: categlist.Name,
				Valid:  true,
			},
			LongName: sql.NullString{
				String: categlist.WOO_LONGNAME,
				Valid:  true,
			},
			IdentRK: categlist.Ident,
			ParentRK: sql.NullInt32{
				Int32: int32(categlist.Parent),
				Valid: true,
			},
			Sync: sql.NullInt32{
				Int32: int32(categlist.WOO_SYNC),
				Valid: true,
			},
		}

		logger.Debug("Проверка игнор-лист")
		for _, ignoreIdent := range cfg.RK7.CateglistIdentIgnore {
			if categlist.ItemIdent == ignoreIdent {
				logger.Debug("Папка в игнор-листе. Обнулить WOO/RK7")
				// todo
				categlistRowDb.Status = sql.NullString{
					String: categlistDatabase.IGNORE,
					Valid:  true,
				}
				err = categlistRowDb.UpdateByIdentRK(db)
				if err != nil {
					return errors.Wrap(err, "failed in UpdateByIdentRK()")
				}
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
						logger.Debugf("RK.WOO_ID=%d && WOO.ID=%d", categlist.WOO_ID, category.ID)
						if categlistName == category.Name && categlist.WOO_ID == category.ID {
							logger.Debug("Папка RK7 совпадает с WOO(свойства Name/LongName/WOO_ID). Обновление в WOO не требуется")
							// todo
							categlistRowDb.Status = sql.NullString{
								String: categlistDatabase.NOT_NEED_UPDATE,
								Valid:  true,
							}
							err = categlistRowDb.UpdateByIdentRK(db)
							if err != nil {
								return errors.Wrap(err, "failed in UpdateByIdentRK()")
							}
						} else {
							logger.Debug("Папка RK7 не совпадает с WOO(свойства Name/LongName/WOO_ID). Обновляем WOO.")
							// todo
							categlistRowDb.Status = sql.NullString{
								String: categlistDatabase.NEED_UPDATE,
								Valid:  true,
							}
							err = categlistRowDb.UpdateByIdentRK(db)
							if err != nil {
								return errors.Wrap(err, "failed in UpdateByIdentRK()")
							}
						}

					} else {
						logger.Debug("Папка не найдена в WOO. Создаем в WOO.")
						// todo
						categlistRowDb.Status = sql.NullString{
							String: categlistDatabase.NOT_FOUND_IN_WOO,
							Valid:  true,
						}
						err = categlistRowDb.UpdateByIdentRK(db)
						if err != nil {
							return errors.Wrap(err, "failed in UpdateByIdentRK()")
						}
					}
				} else {
					logger.Debug("Не указан WOO_ID. Создаем в WOO.")
					// todo
					categlistRowDb.Status = sql.NullString{
						String: categlistDatabase.NOT_WOO_ID,
						Valid:  true,
					}
					err = categlistRowDb.UpdateByIdentRK(db)
					if err != nil {
						return errors.Wrap(err, "failed in UpdateByIdentRK()")
					}
				}
			} else {
				logger.Debug("Папка не активная. Обнуляем WOO/RK7.")
				categlistNotActive++
				// todo
				categlistRowDb.Status = sql.NullString{
					String: categlistDatabase.NOT_ACTIVE,
					Valid:  true,
				}
				err = categlistRowDb.UpdateByIdentRK(db)
				if err != nil {
					return errors.Wrap(err, "failed in UpdateByIdentRK()")
				}
			}
		} else {
			logger.Debug("Синхронизация отключена. Обнуляем в WOO/RK7.")
			// todo
			categlistRowDb.Status = sql.NullString{
				String: categlistDatabase.SYNC_OFF,
				Valid:  true,
			}
			err = categlistRowDb.UpdateByIdentRK(db)
			if err != nil {
				return errors.Wrap(err, "failed in UpdateByIdentRK()")
			}
		}
	}

	return nil
}

// HandlerCateglistDbOneStage 2 этап - обработка DB.Categlist
func HandlerCateglistDbOneStage(db *sqlx.DB) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerCateglistDbOneStage")
	defer logger.Debug("End HandlerCateglistDbOneStage")

	handlers := []handler{
		{categlistDatabase.IGNORE, HandlerIgnore, "<strong>Ошибка обработки папок из игнор-листа</strong>"},                          //+ todo
		{categlistDatabase.SYNC_OFF, HandlerSyncOff, "<strong>Ошибка обработки папок с выключенной синхронизацией</strong>"},         // todo
		{categlistDatabase.NOT_ACTIVE, HandlerNotActive, "<strong>Ошибка обработки не активных папок</strong>"},                      // todo
		{categlistDatabase.NOT_WOO_ID, HandlerNotWooId, "<strong>Ошибка обработки папок без указанного WOO_ID</strong>"},             // todo
		{categlistDatabase.NOT_FOUND_IN_WOO, HandlerNotFoundInWoo, "<strong>Ошибка обработки папок, не найденных в WOO</strong>"},    // todo
		{categlistDatabase.NEED_UPDATE, HandlerNeedUpdate, "<strong>Ошибка обработки при обновлении папок</strong>"},                 // todo
		{categlistDatabase.NOT_NEED_UPDATE, HandlerNotNeedUpdate, "<strong>Ошибка обработки папок, не требующих обновление/strong>"}, // todo
	}

	for _, h := range handlers {
		logger.Debug("=========================================")
		err := h.Do(db)
		if err != nil {
			return errors.Wrap(err, "failed in handler.Do")
		}
	}

	return nil
}

func HandlerIgnore(db *sqlx.DB, status string, message string) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerIgnore")
	defer logger.Debug("End HandlerIgnore")

	var err error
	var categlists []*categlistDatabase.Categlist
	var query string

	logger.Debug("Выполняем поиск записей в таблице Images")
	query = `SELECT * FROM Categlist WHERE Status=$1 ORDER BY IdentRK;`
	err = db.Select(&categlists, query, status)
	logger.Debugf("SELECT:\n%s(%s)", query, status)
	if err != nil {
		return errors.Wrapf(err, "failed SELECT to dbsqlite; query:\n%s(%s)", query, status)
	} else {
		logger.Debugf("Количество полученных строк: %d", len(categlists))
		if len(categlists) > 0 {
			m, err := CateglistNulledInWooAndRK7(categlists)
			if err != nil {
				return errors.Wrap(err, "failed in CateglistNulledInWooAndRK7")
			}
			if len(m) > 0 {
				m = append(m, message)
				telegram.SendMessageToTelegramWithLogError(strings.Join(m, "\n"))
			}
		}
	}
	return nil
}

func HandlerSyncOff(db *sqlx.DB, status string, message string) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerSyncOff")
	defer logger.Debug("End HandlerSyncOff")

	var err error
	var categlists []*categlistDatabase.Categlist
	var query string

	logger.Debug("Выполняем поиск записей в таблице Images")
	query = `SELECT * FROM Categlist WHERE Status=$1 ORDER BY IdentRK;`
	err = db.Select(&categlists, query, status)
	logger.Debugf("SELECT:\n%s(%s)", query, status)
	if err != nil {
		return errors.Wrapf(err, "failed SELECT to dbsqlite; query:\n%s(%s)", query, status)
	} else {
		logger.Debugf("Количество полученных строк: %d", len(categlists))
		if len(categlists) > 0 {
			m, err :=
				CateglistNulledInWooAndRK7(categlists)
			if err != nil {
				return errors.Wrap(err, "failed in CateglistNulledInWooAndRK7")
			}
			if len(m) > 0 {
				m = append(m, message)
				telegram.SendMessageToTelegramWithLogError(strings.Join(m, "\n"))
			}
		}
	}
	return nil
}

func HandlerNotActive(db *sqlx.DB, status string, message string) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerNotActive")
	defer logger.Debug("End HandlerNotActive")

	var err error
	var categlists []*categlistDatabase.Categlist
	var query string

	logger.Debug("Выполняем поиск записей в таблице Images")
	query = `SELECT * FROM Categlist WHERE Status=$1 ORDER BY IdentRK;`
	err = db.Select(&categlists, query, status)
	logger.Debugf("SELECT:\n%s(%s)", query, status)
	if err != nil {
		return errors.Wrapf(err, "failed SELECT to dbsqlite; query:\n%s(%s)", query, status)
	} else {
		logger.Debugf("Количество полученных строк: %d", len(categlists))
		if len(categlists) > 0 {
			m, err := CateglistNulledInWooAndRK7(categlists)
			if err != nil {
				return errors.Wrap(err, "failed in CateglistNulledInWooAndRK7")
			}
			if len(m) > 0 {
				m = append(m, message)
				telegram.SendMessageToTelegramWithLogError(strings.Join(m, "\n"))
			}
		}
	}
	return nil
}

func HandlerNotWooId(db *sqlx.DB, status string, message string) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerNotWooId")
	defer logger.Debug("End HandlerNotWooId")

	var err error
	var categlists []*categlistDatabase.Categlist
	var query string

	logger.Debug("Выполняем поиск записей в таблице Images")
	query = `SELECT * FROM Categlist WHERE Status=$1 ORDER BY IdentRK;`
	err = db.Select(&categlists, query, status)
	logger.Debugf("SELECT:\n%s(%s)", query, status)
	if err != nil {
		return errors.Wrapf(err, "failed SELECT to dbsqlite; query:\n%s(%s)", query, status)
	} else {
		logger.Debugf("Количество полученных строк: %d", len(categlists))
		if len(categlists) > 0 {
			m, err := CateglistCreateInWooAndRK7(categlists)
			if err != nil {
				return errors.Wrap(err, "failed in CateglistCreateInWooAndRK7")
			}
			if len(m) > 0 {
				m = append(m, message)
				telegram.SendMessageToTelegramWithLogError(strings.Join(m, "\n"))
			}
		}
	}
	return nil
}

func HandlerNotFoundInWoo(db *sqlx.DB, status string, message string) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerNotWooId")
	defer logger.Debug("End HandlerNotWooId")

	var err error
	var categlists []*categlistDatabase.Categlist
	var query string

	logger.Debug("Выполняем поиск записей в таблице Images")
	query = `SELECT * FROM Categlist WHERE Status=$1 ORDER BY IdentRK;`
	err = db.Select(&categlists, query, status)
	logger.Debugf("SELECT:\n%s(%s)", query, status)
	if err != nil {
		return errors.Wrapf(err, "failed SELECT to dbsqlite; query:\n%s(%s)", query, status)
	} else {
		logger.Debugf("Количество полученных строк: %d", len(categlists))
		if len(categlists) > 0 {
			m, err := CateglistCreateInWooAndRK7(categlists)
			if err != nil {
				return errors.Wrap(err, "failed in CateglistCreateInWooAndRK7")
			}
			if len(m) > 0 {
				m = append(m, message)
				telegram.SendMessageToTelegramWithLogError(strings.Join(m, "\n"))
			}
		}
	}
	return nil
}

func HandlerNeedUpdate(db *sqlx.DB, status string, message string) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerNeedUpdate")
	defer logger.Debug("End HandlerNeedUpdate")

	var err error
	var categlists []*categlistDatabase.Categlist
	var query string

	logger.Debug("Выполняем поиск записей в таблице Images")
	query = `SELECT * FROM Categlist WHERE Status=$1 ORDER BY IdentRK;`
	err = db.Select(&categlists, query, status)
	logger.Debugf("SELECT:\n%s(%s)", query, status)
	if err != nil {
		return errors.Wrapf(err, "failed SELECT to dbsqlite; query:\n%s(%s)", query, status)
	} else {
		logger.Debugf("Количество полученных строк: %d", len(categlists))
		if len(categlists) > 0 {
			m, err := CateglistUpdateInWoo(categlists)
			if err != nil {
				return errors.Wrap(err, "failed in CateglistUpdateInWoo")
			}
			if len(m) > 0 {
				m = append(m, message)
				telegram.SendMessageToTelegramWithLogError(strings.Join(m, "\n"))
			}
		}
	}
	return nil
}

func HandlerNotNeedUpdate(db *sqlx.DB, status string, message string) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerNotNeedUpdate")
	defer logger.Debug("End HandlerNotNeedUpdate")

	return nil
}

// HandlerCateglistUpdateParentId 3 этап - синхронизация DB.Categlist.Parent и WOO.ProductCategory.Parent
func HandlerCateglistUpdateParentId(db *sqlx.DB) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerCateglistUpdateParentId")
	defer logger.Debug("End HandlerCateglistUpdateParentId")

	handlers := []handler{
		{categlistDatabase.NOT_NEED_UPDATE, HandlerNeedUpdateParentId, "<strong>Ошибка при обновлении ParentID в RK7/WOO/strong>"}, // todo
	}

	for _, h := range handlers {
		logger.Debug("=========================================")
		err := h.Do(db)
		if err != nil {
			return errors.Wrap(err, "failed in handler.Do")
		}
	}

	return nil
}

func HandlerNeedUpdateParentId(db *sqlx.DB, status string, message string) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerNeedUpdate")
	defer logger.Debug("End HandlerNeedUpdate")

	var err error
	var categlists []*categlistDatabase.Categlist
	var query string

	logger.Debug("Выполняем поиск записей в таблице Images")
	query = `SELECT * FROM Categlist WHERE Status=$1 ORDER BY IdentRK;`
	err = db.Select(&categlists, query, status)
	logger.Debugf("SELECT:\n%s(%s)", query, status)
	if err != nil {
		return errors.Wrapf(err, "failed SELECT to dbsqlite; query:\n%s(%s)", query, status)
	} else {
		logger.Debugf("Количество полученных строк: %d", len(categlists))
		if len(categlists) > 0 {
			m, err := CateglistUpdateParentIdInRk(categlists)
			if err != nil {
				return errors.Wrap(err, "failed in CateglistUpdateParentIdInRk")
			}
			if len(m) > 0 {
				m = append(m, message)
				telegram.SendMessageToTelegramWithLogError(strings.Join(m, "\n"))
			}
		}
	}
	return nil
}

/////////////////////////

// CateglistNulledInWooAndRK7 - удалить папку в WOO с предварительным поиском и обнулить в RK7.
// Используется во 2 этапе синхронизации папок
func CateglistNulledInWooAndRK7(categlistsInDb []*categlistDatabase.Categlist) ([]string, error) {
	logger := logging.GetLogger()
	logger.Debug("Start CateglistNulledInWooAndRK7")
	defer logger.Debug("End CateglistNulledInWooAndRK7")

	var m []string
	var err error

	menu, err := cache.GetMenu()
	if err != nil {
		return nil, errors.Wrap(err, "failed in cache.GetMenu()")
	}

	categlistsRK7ByIdent, err := menu.GetCateglistsRK7ByIdent()
	if err != nil {
		return nil, errors.Wrap(err, "failed in GetCateglistsRK7ByIdent()")
	}

	productCategoriesWooByID, err := menu.GetProductCategoriesWooByID()
	if err != nil {
		return nil, errors.Wrap(err, "failed in GetProductCategoriesWooByID()")
	}

	woo := wooapi.GetAPI()
	rk7 := rk7api.GetAPI("REF")

	for _, categlistInDb := range categlistsInDb {
		if categlistInRk7, ok := categlistsRK7ByIdent[categlistInDb.IdentRK]; ok {
			logger.Debugf("Папка %s", GetCateglistDescription(categlistInRk7))

			logger.Debug("Пробуем найти и удалить в WOO")
			if categlistInDb.IdentWOO.Valid {
				if categlistInDb.IdentWOO.Int32 != 0 {
					if productCategory, ok := productCategoriesWooByID[int(categlistInDb.IdentWOO.Int32)]; ok {
						logger.Debugf("ProductCategory %s", GetProductCategoryDescription(productCategory))
						err := woo.ProductCategoryDelete(productCategory.ID, options.Force(true))
						if err != nil {
							m = append(m, fmt.Sprintf("%s; Не удалось удалить папку в WOO; %v", GetCateglistDescription(categlistInRk7), err))
						} else {
							logger.Debug("ProductCategory успешно удален")
						}
					} else {
						logger.Debugf("ProductCategory(id=%d) не найден в WOO", categlistInDb.IdentWOO.Int32)
					}
				} else {
					logger.Debug("WOO_ID = 0")
				}
			} else {
				logger.Debug("Нет указан WOO_ID")
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
					m = append(m, fmt.Sprintf("%s; Не удалось обнулить папку в RK7; %v", GetCateglistDescription(categlistInRk7), err))
				} else {
					logger.Debug("Папка успешно удалена в RK7")
				}
			} else {
				logger.Debug("Обнуление в RK7 не требуется. WOO_ID/WOO_PARENT_ID = 0")
			}
		} else {
			m = append(m, fmt.Sprintf("Папка(db=%v) не найдена в RK7", categlistInDb))
		}
	}
	if len(m) > 0 {
		return m, nil
	} else {
		return nil, nil
	}

}

// CateglistCreateInWooAndRK7 - создать папку в WOO и обновить WOO_ID в RK7.
// Используется во 2 этапе синхронизации папок
func CateglistCreateInWooAndRK7(categlistsInDb []*categlistDatabase.Categlist) ([]string, error) {
	logger := logging.GetLogger()
	logger.Debug("Start CateglistCreateInWooAndRK7")
	defer logger.Debug("End CateglistCreateInWooAndRK7")

	var m []string
	var err error

	menu, err := cache.GetMenu()
	if err != nil {
		return nil, errors.Wrap(err, "failed in cache.GetMenu()")
	}

	categlistsRK7ByIdent, err := menu.GetCateglistsRK7ByIdent()
	if err != nil {
		return nil, errors.Wrap(err, "failed in GetCateglistsRK7ByIdent()")
	}

	woo := wooapi.GetAPI()
	rk7 := rk7api.GetAPI("REF")
	cfg := config.GetConfig()

	for _, categlistInDb := range categlistsInDb {
		if categlistInRk7, ok := categlistsRK7ByIdent[categlistInDb.IdentRK]; ok {
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
			categoryCreated, err := woo.ProductCategoryAdd(category)
			if err != nil {
				if err.Error() == "code:term_exists; message:Элемент с указанным именем уже существует у родительского элемента.; status:400; display:; details:;" {
					category.Name = fmt.Sprintf("%s_%d", categlistName, categlistInRk7.Ident)
					categoryCreated, err = woo.ProductCategoryAdd(category)
					if err != nil {
						m = append(m, fmt.Sprintf("failed in ProductCategoryAdd(%v); %v", category, err))
					}
				} else {
					m = append(m, fmt.Sprintf("failed in ProductCategoryAdd(%v); %v", category, err))
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
					m = append(m, fmt.Sprintf("Ошибка при добавление папки в кеш WOO; %v", err))
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
						m = append(m, fmt.Sprintf("Ошибка при обновлении WOO_ID/WOO_PARENT_ID в RK7. Кеш установлен по умолчанию; %v", err))
					} else {
						logger.Debug("Папка успешно обновлена")
					}
				}
			} else {
				m = append(m, fmt.Sprintf("Не удалось создать папку в WOO; ProductCategoryAdd(Name=%d)", categlistInRk7.Name))
			}
		} else {
			m = append(m, fmt.Sprintf("Папка(db=%v) не найдена в RK7", categlistInDb))
		}
	}
	if len(m) > 0 {
		return m, nil
	} else {
		return nil, nil
	}
}

// CateglistUpdateInWoo - обновить папку в WOO и обновить WOO_ID в RK7.
// Используется во 2 этапе синхронизации папок
func CateglistUpdateInWoo(categlistsInDb []*categlistDatabase.Categlist) ([]string, error) {
	logger := logging.GetLogger()
	logger.Debug("Start CateglistUpdateInWoo")
	defer logger.Debug("End CateglistUpdateInWoo")

	var m []string
	var err error

	menu, err := cache.GetMenu()
	if err != nil {
		return nil, errors.Wrap(err, "failed in cache.GetMenu()")
	}

	categlistsRK7ByIdent, err := menu.GetCateglistsRK7ByIdent()
	if err != nil {
		return nil, errors.Wrap(err, "failed in GetCateglistsRK7ByIdent()")
	}

	woo := wooapi.GetAPI()

	for _, categlistInDb := range categlistsInDb {
		if categlistInRk7, ok := categlistsRK7ByIdent[categlistInDb.IdentRK]; ok {

			logger.Debugf("Обновляем папку в WOO/кеше WOO")
			category, err := woo.ProductCategoryGet(categlistInRk7.WOO_ID)
			if err != nil {
				m = append(m, fmt.Sprintf("failed in ProductCategoryGet(ID=%d); %v", categlistInRk7.WOO_ID, err))
			} else {
				if category != nil {
					logger.Debugf("ProductCategory успешно получен: Name=%s, ID=%d, Parent=%d, Slug=%s", category.Name, category.ID, category.Parent, category.Slug)

					var categlistName string
					if categlistInRk7.WOO_LONGNAME != "" {
						categlistName = categlistInRk7.WOO_LONGNAME
					} else {
						categlistName = categlistInRk7.Name
					}

					recoveryName := category.Name
					if category.Name != categlistName {
						category.Name = categlistName
					}
					recoveryParent := category.Parent // todo надо ли выполнять обновление parentID?
					if category.Parent != categlistInRk7.WOO_PARENT_ID {
						category.Parent = categlistInRk7.WOO_PARENT_ID
					}

					_, err = woo.ProductCategoryUpdate(category)
					if err != nil {
						category.Name = recoveryName
						category.Parent = recoveryParent
						m = append(m, fmt.Sprintf("failed in ProductCategoryUpdate(ID=%d, Name=%s); %v", category.ID, category.Name, err))
					} else {
						logger.Debug("Папка успешно обновлена. Кеш обновлен")
					}
				} else {
					m = append(m, fmt.Sprintf("failed in ProductCategoryGet(ID=%d)", categlistInRk7.WOO_ID))
				}
			}
		} else {
			m = append(m, fmt.Sprintf("Папка(db=%v) не найдена в RK7", categlistInDb))
		}
	}
	if len(m) > 0 {
		return m, nil
	} else {
		return nil, nil
	}
}

//////////////////

// CateglistUpdateParentIdInRk - обновить ParentID в RK7 и в случае успеха обновить в WOO
// Используется во 3 этапе синхронизации папок
func CateglistUpdateParentIdInRk(categlistsInDb []*categlistDatabase.Categlist) ([]string, error) {
	logger := logging.GetLogger()
	logger.Debug("Start CateglistUpdateParentIdInRk")
	defer logger.Debug("End CateglistUpdateParentIdInRk")

	var m []string
	var err error

	menu, err := cache.GetMenu()
	if err != nil {
		return nil, errors.Wrap(err, "failed in cache.GetMenu()")
	}

	categlistsRK7ByIdent, err := menu.GetCateglistsRK7ByIdent()
	if err != nil {
		return nil, errors.Wrap(err, "failed in GetCateglistsRK7ByIdent()")
	}

	rk7 := rk7api.GetAPI("REF")
	cfg := config.GetConfig()

	for _, categlistInDb := range categlistsInDb {
		if categlistInRk7, ok := categlistsRK7ByIdent[categlistInDb.IdentRK]; ok {
			logger.Debug("=========================================")
			logger.Debugf("Папка %s", GetCateglistDescription(categlistInRk7))

			// получаем parentID
			// сравниваем
			// вставляем если надо

			if categlistInRk7Parent, found := categlistsRK7ByIdent[categlistInRk7.MainParentIdent]; found {
				logger.Debug("Папка Parent найдена в кеше RK7")
				var parentID int
				switch {
				case categlistInRk7Parent.ItemIdent == 0:
					logger.Debug("Папка Parent корневая - используем WOO_ID из cfg.WOOCOMMERCE.MenuCategoryId")
					parentID = cfg.WOOCOMMERCE.MenuCategoryId
				case categlistInRk7Parent.WOO_SYNC != 1:
					logger.Debug("Папка Parent с выключенной синхронизацией - используем WOO_ID из cfg.WOOCOMMERCE.MenuCategoryId")
					parentID = cfg.WOOCOMMERCE.MenuCategoryId
				default:
					logger.Debug("Папка Parent не корневая - используем WOO_ID из categlistsRK7ByIdent[categlist.MainParentIdent]")
					parentID = categlistInRk7Parent.WOO_ID
				}

				if categlistInRk7.WOO_PARENT_ID != parentID {
					logger.Debug("Обновляем WOO_PARENT_ID в RK7/кеше RK7")
					var categlists []*models.Categlist
					recoveryWooParentID := categlistInRk7.WOO_PARENT_ID
					categlistInRk7.WOO_PARENT_ID = parentID
					categlists = append(categlists, categlistInRk7)
					_, err = rk7.SetRefDataCateglist(categlists)
					if err != nil {
						categlistInRk7.WOO_PARENT_ID = recoveryWooParentID
						m = append(m, fmt.Sprintf("Ошибка при обновлении WOO_PARENT_ID в RK7. Кеш установлен по умолчанию; %v", err))
					} else {
						logger.Debugf("Папка успешно обновлена в RK7/кеше RK7. ParentID = %d", parentID)
						// todo сделать чтобы было обновление в WOO или оаствить обычные проверки как есть
						logger.Info("Обновляем ParentID в WOO")
						err := UpdateCateglistInWoo(categlistInRk7)
						if err != nil {
							m = append(m, fmt.Sprintf("Ошибка при обновлении WOO_PARENT_ID в WOO; %v", err))
						} else {
							logger.Debug("ParentID в WOO обновлен")
						}
					}
				} else {
					logger.Debug("Обновление WOO_PARENT_ID в RK7 не требуется")
				}
			} else {
				m = append(m, fmt.Sprintf("Папка Parent(ID=%d) не найдена в RK7", categlistInRk7.MainParentIdent))
			}
		} else {
			m = append(m, fmt.Sprintf("Папка(db=%v) не найдена в RK7", categlistInDb))
		}
	}
	if len(m) > 0 {
		return m, nil
	} else {
		return nil, nil
	}

}

func UpdateCateglistInWoo(categlist *models.Categlist) error {
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
