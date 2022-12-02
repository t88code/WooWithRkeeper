package sync

import (
	"WooWithRkeeper/internal/cache"
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/internal/database"
	"WooWithRkeeper/internal/database/model"
	"database/sql"

	"WooWithRkeeper/pkg/logging"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"os"
	"time"
)

// обработка Menuitems и запись результатов в DB.Image
func HandlerMenuitemsToDbImage() error {

	logger := logging.GetLogger()
	logger.Debug("Start HandlerMenuitemsToDbImage")
	defer logger.Debug("End HandlerMenuitemsToDbImage")
	var err error
	cfg := config.GetConfig()

	db, err := sqlx.Connect("sqlite3", database.DB_NAME)
	if err != nil {
		return errors.Wrap(err, "failed sqlx.Connect")
	}
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			logger.Fatalf("failed close sqlx.Connect, err: %v", err)
		}
	}(db)

	logger.Debug("Получаем меню из RK7 и WOO")
	menu, err := cache.GetMenu()
	if err != nil {
		return errors.Wrap(err, "failed in cache.GetMenu()")
	}

	menuitems, err := menu.GetMenuitems()
	if err != nil {
		return errors.Wrap(err, "failed in menu.GetMenuitems()")
	}

	productsWooByID, err := menu.GetProductsWooByID()
	if err != nil {
		return errors.Wrap(err, "failed in menu.GetProductsWooByID()")
	}

	dishRestsByIdent, err := menu.GetDishRestsByIdent()
	if err != nil {
		return errors.Wrap(err, "failed in menu.GetDishRestsByIdent()")
	}

	// блюда RK7
	var menuitemsActive int    // активные - счетчик
	var menuitemsNotActive int // не активные - счетчик

	var menuitemsFoundInWoo int    // найдены в WOO
	var menuitemsNotFoundInWoo int // не найдены в WOO

LoopOneStage:
	for i, menuitem := range menuitems {
		logger.Debug("--------------------------------------")
		imageNames := menuitems[i].GetImageNames()

		dish := fmt.Sprintf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d, RK_Image: %v",
			menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES, imageNames)
		logger.Debugf(dish)

		imageRowDb := model.Image{IdentRK: menuitem.Ident}

		logger.Debug("Проверка игнор-лист")
		for _, ignoreIdent := range cfg.RK7.MenuitemIdentIgnore {
			if menuitem.ItemIdent == ignoreIdent {
				logger.Warning("Блюдо в игнор-листе")
				imageRowDb.Status = sql.NullString{
					String: model.IMAGE_STATUS_IGNORE,
					Valid:  true,
				}
				err = imageRowDb.UpdateByIdentRKAndPos(db)
				if err != nil {
					return errors.Wrap(err, "failed in UpdateByIdentRKAndPos()")
				}
				continue LoopOneStage
			}
		}

		if menuitem.Status != 3 {
			logger.Debug("Не активное блюдо. Пропускаем. Синхронизация меню отключила/удалила блюдо в WOO")
			menuitemsNotActive++
			imageRowDb.Status = sql.NullString{
				String: model.IMAGE_STATUS_IGNORE,
				Valid:  true,
			}
			err = imageRowDb.UpdateByIdentRKAndPos(db)
			if err != nil {
				return errors.Wrap(err, "failed in UpdateByIdentRKAndPos()")
			}
		} else {
			logger.Debug("Блюдо активное. Проверяем цену/стоп-лист/включение синхронизации")
			menuitemsActive++
			logger.Debug("Проверяем наличие в стоп-листе")
			if dishRests, foundInStopList := dishRestsByIdent[menuitem.Ident]; foundInStopList {
				if dishRests.Prohibited == 1 || dishRests.Quantity == 0 {
					logger.Warning("Блюдо в стоп-листе")
					imageRowDb.Status = sql.NullString{
						String: model.IMAGE_STATUS_IGNORE,
						Valid:  true,
					}
					err = imageRowDb.UpdateByIdentRKAndPos(db)
					if err != nil {
						return errors.Wrap(err, "failed in UpdateByIdentRKAndPos()")
					}
					continue LoopOneStage
				}
			}

			if menuitem.PRICETYPES == 9223372036854775807 {
				logger.Warning("Блюдо - цена не указана")
				imageRowDb.Status = sql.NullString{
					String: model.IMAGE_STATUS_IGNORE,
					Valid:  true,
				}
				err = imageRowDb.UpdateByIdentRKAndPos(db)
				if err != nil {
					return errors.Wrap(err, "failed in UpdateByIdentRKAndPos()")
				}
				continue LoopOneStage
			}

			if menuitem.CLASSIFICATORGROUPS != cfg.RK7.CLASSIFICATORGROUPSALLOW {
				logger.Warning("Блюдо - с выключенной синхронизацией	")
				imageRowDb.Status = sql.NullString{
					String: model.IMAGE_STATUS_IGNORE,
					Valid:  true,
				}
				err = imageRowDb.UpdateByIdentRKAndPos(db)
				if err != nil {
					return errors.Wrap(err, "failed in UpdateByIdentRKAndPos()")
				}
				continue LoopOneStage
			}

			if _, found := productsWooByID[menuitem.WOO_ID]; found {
				logger.Debug("Блюдо найдено в WOO")
				menuitemsFoundInWoo++
				logger.Debugf("Всего картинок %d штук", len(imageNames))
				for imageNameIndex, imageName := range imageNames {
					imageRowDb := model.Image{IdentRK: menuitem.Ident}
					imageRowDb.Pos = sql.NullInt32{
						Int32: int32(imageNameIndex),
						Valid: true,
					}
					if imageName == "" {
						logger.Warning("Не указано наименование картинки. Обнулить картинки в WOO")
						imageRowDb.Status = sql.NullString{
							String: model.IMAGE_STATUS_RK7_IMAGE_NAME_NOT_FOUND,
							Valid:  true,
						}
						err := imageRowDb.UpdateByIdentRKAndPos(db)
						if err != nil {
							return err
						}
					} else {
						logger.Debugf("Указано наименование картинки %s, позиция %d", imageName, imageNameIndex)
						logger.Debugf("Проверяем наличие картинки в папке %s", cfg.IMAGESYNC.Path)
						imageRowDb.Name = sql.NullString{
							String: imageName,
							Valid:  true,
						}
						path := fmt.Sprintf("%s/%s.jpg", cfg.IMAGESYNC.Path, imageName)
						if fileInfo, err := os.Stat(path); !os.IsNotExist(err) {
							logger.Debug("Получаем запись из DB")
							var imageInDb []model.Image
							query := "SELECT * FROM Image WHERE IdentRK = ? and Pos = ?"
							err = db.Select(&imageInDb, query, menuitem.ItemIdent, imageNameIndex)
							if err != nil {
								return errors.Wrapf(err, "failed SELECT to dbsqlite; dish %s; query %s(%d, %d)", dish, query, menuitem.ItemIdent, imageNameIndex)
							} else {
								switch {
								case len(imageInDb) == 0:
									logger.Warning("Запись не найдена в DB. Необходимо создать запись в DB и обновить картинку в WOO")
									imageRowDb.Status = sql.NullString{
										String: model.IMAGE_STATUS_NEED_UPDATE_BY_NOT_FOUND_IN_DB,
										Valid:  true,
									}
									err := imageRowDb.UpdateByIdentRKAndPos(db)
									if err != nil {
										return err
									}
								case len(imageInDb) == 1:
									logger.Debug("Блюдо найдено в DB. Сверяем имя/дату изменения")
									logger.Debugf("Наименование картинки: DB=%s, RK7=%s, File=%s", imageInDb[0].Name, imageName, fileInfo.Name())
									logger.Debugf("Дата изменения картинки: DB=%s, File=%s", imageInDb[0].ModTime.String, fileInfo.ModTime().Format(time.RFC3339))
									switch {
									case imageInDb[0].Name.String != imageName:
										logger.Debug("Наименование картинки изменилось. Обновляем картинку в WOO")
										imageRowDb.Status = sql.NullString{
											String: model.IMAGE_STATUS_NEED_UPDATE_BY_DIFF_NAME,
											Valid:  true,
										}
										err := imageRowDb.UpdateByIdentRKAndPos(db)
										if err != nil {
											return err
										}
									case imageInDb[0].ModTime.String != fileInfo.ModTime().Format(time.RFC3339):
										logger.Debug("Дата изменения картинки изменилась. Необходимо обновить картинку")
										imageRowDb.Status = sql.NullString{
											String: model.IMAGE_STATUS_NEED_UPDATE_BY_DIFF_DATE,
											Valid:  true,
										}
										err := imageRowDb.UpdateByIdentRKAndPos(db)
										if err != nil {
											return err
										}
									default:
										logger.Debug("Дата изменения картинки не изменилась. Наименование не изменилось. Необходимо проверить наличие в WOO")
										imageRowDb.Status = sql.NullString{
											String: model.IMAGE_STATUS_NO_NEED_UPDATE,
											Valid:  true,
										}
										err := imageRowDb.UpdateByIdentRKAndPos(db)
										if err != nil {
											return err
										}
									}
								case len(imageInDb) > 1:
									logger.Errorf("Недопустимое количество блюд в DB = %d > 1", len(imageInDb))
									imageRowDb.Status = sql.NullString{
										String: model.IMAGE_STATUS_NEED_UPDATE_BY_FIND_DOUBLE_IN_DB,
										Valid:  true,
									}
									err := imageRowDb.UpdateByIdentRKAndPos(db)
									if err != nil {
										return err
									}
								default:
									return errors.New("Неизвестная ошибка")
								}
							}
						} else {
							logger.Error("Файл картинки не найден. Отправить сообщение и обнулить картинку в WOO")
							imageRowDb.Status = sql.NullString{
								String: model.IMAGE_STATUS_FILE_NOT_FOUND,
								Valid:  true,
							}
							err := imageRowDb.UpdateByIdentRKAndPos(db)
							if err != nil {
								return err
							}
						}
					}

				}
			} else {
				menuitemsNotFoundInWoo++
				if menuitem.WOO_ID != 0 {
					logger.Error("Блюдо не найдено в WOO. Отправить сообщение")
					imageRowDb.Status = sql.NullString{
						String: model.IMAGE_STATUS_WOO_NOT_FOUND,
						Valid:  true,
					}
					err = imageRowDb.UpdateByIdentRKAndPos(db)
					if err != nil {
						return errors.Wrap(err, "failed in UpdateByIdentRKAndPos()")
					}
				} else {
					logger.Warning("Блюдо без WOO_ID. Игнорировать")
					imageRowDb.Status = sql.NullString{
						String: model.IMAGE_STATUS_RK7_WOO_ID_NOT_FOUND,
						Valid:  true,
					}
					err = imageRowDb.UpdateByIdentRKAndPos(db)
					if err != nil {
						return errors.Wrap(err, "failed in UpdateByIdentRKAndPos()")
					}
				}
			}
		}
	}

	return nil
}

// обработка DB.Image
func HandlerDbImage() error {

	logger := logging.GetLogger()
	logger.Debug("Start HandlerDbImage")
	defer logger.Debug("End HandlerDbImage")

	handlers := []handler{
		{model.IMAGE_STATUS_WOO_NOT_FOUND, HandlerWooNotFound},
		{model.IMAGE_STATUS_RK7_IMAGE_NAME_NOT_FOUND, HandlerRk7ImageNameNotFound},
		{model.IMAGE_STATUS_FILE_NOT_FOUND, HandlerFileNotFound},
		{model.IMAGE_STATUS_NEED_UPDATE_BY_DIFF_NAME, HandlerNeedUpdateByDiffName},
		{model.IMAGE_STATUS_NEED_UPDATE_BY_DIFF_DATE, HandlerNeedUpdateByDiffDate},
		{model.IMAGE_STATUS_NEED_UPDATE_BY_NOT_FOUND_IN_DB, HandlerNeedUpdateByNotFoundInDb},
		{model.IMAGE_STATUS_NEED_UPDATE_BY_FIND_DOUBLE_IN_DB, HandlerNeedUpdateByFindDoubleInDb},
		{model.IMAGE_STATUS_NO_NEED_UPDATE, HandlerNoNeedUpdate},
	}

	for _, h := range handlers {
		logger.Error("=========================================")
		err := h.Do()
		if err != nil {
			return errors.Wrap(err, "failed in handler.Do")
		}
	}

	return nil
}

type handler struct {
	status  string
	handler func(*sqlx.DB, []*model.Image) error
}

func (h *handler) Do() error {
	logger := logging.GetLogger()
	logger.Debugf("Start doHandler.%s", h.status)
	defer logger.Debugf("Start doHandler.%s", h.status)

	db, err := sqlx.Connect("sqlite3", database.DB_NAME)
	if err != nil {
		return errors.Wrap(err, "failed sqlx.Connect")
	}
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			logger.Fatalf("failed close sqlx.Connect, err: %v", err)
		}
	}(db)

	i := new(model.Image)
	i.Status = sql.NullString{
		String: h.status,
		Valid:  true,
	}
	images, err := i.SelectByStatus(db)
	if err != nil {
		return errors.Wrap(err, "failed in SelectByIdentRKAndPos")
	}
	return errors.Wrapf(h.handler(db, images), "failed in handler(%s)", h.status)
}

//IMAGE_STATUS_WOO_NOT_FOUND
func HandlerWooNotFound(db *sqlx.DB, images []*model.Image) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerWooNotFound")
	defer logger.Debug("End HandlerWooNotFound")

	for _, image := range images {
		logger.Warning(image)
	}
	return nil
}

//IMAGE_STATUS_RK7_IMAGE_NAME_NOT_FOUND
func HandlerRk7ImageNameNotFound(db *sqlx.DB, images []*model.Image) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerRk7ImageNameNotFound")
	defer logger.Debug("End HandlerRk7ImageNameNotFound")

	for _, image := range images {
		logger.Warning(image)
	}
	return nil
}

//IMAGE_STATUS_FILE_NOT_FOUND
func HandlerFileNotFound(db *sqlx.DB, images []*model.Image) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerFileNotFound")
	defer logger.Debug("End HandlerFileNotFound")

	for _, image := range images {
		logger.Warning(image)
	}
	return nil
}

//IMAGE_STATUS_NEED_UPDATE_BY_DIFF_NAME
func HandlerNeedUpdateByDiffName(db *sqlx.DB, images []*model.Image) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerNeedUpdateByDiffName")
	defer logger.Debug("End HandlerNeedUpdateByDiffName")

	for _, image := range images {
		logger.Warning(image)
	}
	return nil
}

//IMAGE_STATUS_NEED_UPDATE_BY_DIFF_DATE
func HandlerNeedUpdateByDiffDate(db *sqlx.DB, images []*model.Image) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerNeedUpdateByDiffDate")
	defer logger.Debug("End HandlerNeedUpdateByDiffDate")

	for _, image := range images {
		logger.Warning(image)
	}
	return nil
}

//IMAGE_STATUS_NEED_UPDATE_BY_NOT_FOUND_IN_DB
func HandlerNeedUpdateByNotFoundInDb(db *sqlx.DB, images []*model.Image) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerNeedUpdateByNotFoundInDb")
	defer logger.Debug("End HandlerNeedUpdateByNotFoundInDb")

	for _, image := range images {
		logger.Warning(image)
	}
	return nil
}

//IMAGE_STATUS_NEED_UPDATE_BY_FIND_DOUBLE_IN_DB
func HandlerNeedUpdateByFindDoubleInDb(db *sqlx.DB, images []*model.Image) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerNeedUpdateByFindDoubleInDb")
	defer logger.Debug("End HandlerNeedUpdateByFindDoubleInDb")

	for _, image := range images {
		logger.Warning(image)
	}
	return nil
}

//IMAGE_STATUS_NO_NEED_UPDATE
func HandlerNoNeedUpdate(db *sqlx.DB, images []*model.Image) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerNoNeedUpdate")
	defer logger.Debug("End HandlerNoNeedUpdate")

	for _, image := range images {
		logger.Warning(image)
	}
	return nil
}

func SyncImages() error {
	logger := logging.GetLogger()
	logger.Debug("Start SyncImages")
	defer logger.Debug("End SyncImages")
	var err error

	err = HandlerMenuitemsToDbImage()
	if err != nil {
		return errors.Wrap(err, "failed in HandlerMenuitemsToDbImage")
	}

	err = HandlerDbImage()
	if err != nil {
		return errors.Wrap(err, "failed in HandlerDbImage")
	}

	return nil
}

func SyncImages1() error {

	logger := logging.GetLogger()
	logger.Debug("Start SyncImages")
	defer logger.Debug("End SyncImages")
	var err error
	cfg := config.GetConfig()

	db, err := sqlx.Connect("sqlite3", database.DB_NAME)
	if err != nil {
		logger.Fatalf("failed sqlx.Connect; %v", err)
	}
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			logger.Fatalf("failed close sqlx.Connect, err: %v", err)
		}
	}(db)

	//var resultSyncAll []string
	//var resultSyncError []string

	logger.Debug("Получаем меню из RK7 и WOO")
	menu, err := cache.GetMenu()
	if err != nil {
		return err
	}

	menuitems, err := menu.GetMenuitems()
	if err != nil {
		return err
	}

	productsWooByID, err := menu.GetProductsWooByID()
	if err != nil {
		return err
	}

	dishRestsByIdent, err := menu.GetDishRestsByIdent()
	if err != nil {
		return err
	}

	logger.Debug("Запущен процесс обновления картинок блюд")

	// блюда RK7
	var menuitemsActive int    // активные - счетчик
	var menuitemsNotActive int // не активные - счетчик

	//// игнор
	//var menuitemsPriceNotDefineStopListNoneSync []*modelsRK7API.MenuitemItem //+ не указана цена/стоп-лист/игнор/выключена sync через классификацию - do Отправить сообщение - todo(28.11.22) Проверить, что картинок нет в WOO и DB.Status=Ignore
	//
	//// не найден в WOO
	//var menuitemsNotFoundInWoo []*modelsRK7API.MenuitemItem //+ не найдено в WOO - do Отправить сообщение. Ничего не делаем, потому что основная синхронизация должна решить этот вопрос - todo(28.11.22) Сообщение об ошибке; DB.Status=InWooNotFound
	//var menuitemsWooIsNull []*modelsRK7API.MenuitemItem     //+ WooID не указан - do Отправить сообщение - todo(28.11.22) DB.Status=Rk7WooIDNotFound

	// найден в WOO, с ошибками настройки
	var menuitemsFoundInWoo int // найдено в WOO - счетчик
	//var menuitemsImageNameNull []*modelsRK7API.MenuitemItem     //+ WooImageName не указан - do Отправить сообщение и обнулить картинку в WOO --todo(28.11.22) Обнулить все картинки в WOO и DB.Status=Rk7ImageNameNotFound
	//var menuitemsImageFileNotFound []*modelsRK7API.MenuitemItem //+ WooImageName указан, но файл картинки не найден - do Отправить сообщение и обнулить картинку в WOO --todo(28.11.22) Сообщить об ошибке, обнулить все картинки в WOO и DB.Status=ImageFileNotFound
	//
	//// найден в WOO, картинка найдена
	//var menuitemsNeedUpdateInWooByName []*modelsRK7API.MenuitemItem       //+ наименование картинки изменилась - do Обновить/создать картинку в WOO --todo(28.11.22) Обновить совместно все картинки и DB.Status=NeedUpdateByDiffName
	//var menuitemsNeedUpdateInWooByDate []*modelsRK7API.MenuitemItem       //+ дата картинки изменилась - do Обновить/создать картинку в WOO --todo(28.11.22) Обновить совместно все картинки и DB.Status=NeedUpdateByDiff // Date
	//var menuitemsNoNeedUpdateNeedVerifyInWoo []*modelsRK7API.MenuitemItem //+ дата картинки не изменилась, наименование совпадает - do Проверить наличие картинки в WOO --todo(28.11.22) Если других обновлений нет, то сверить с WOO и DB.Status=NoNeedUpdate
	//var menuitemsDubleInDB []*modelsRK7API.MenuitemItem                   //+ дубли в DB - do Отправить сообщение --todo(28.11.22) Обновить совместно все картинки и DB.Status=NeedUpdateByDoubleInDb
	//var menuitemsNeedAddInDB []*modelsRK7API.MenuitemItem                 //+ нет записи в DB - do Обновить/создать картинку в WOO и добавить запись в DB --todo(28.11.22) Обновить совместно все картинки и DB.Status=NeedUpdateByNotFoundInDb

LoopOneStage:
	for i, menuitem := range menuitems {
		logger.Debug("--------------------------------------")
		imageNames := menuitems[i].GetImageNames()

		dish := fmt.Sprintf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d, RK_Image: %v",
			menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES, imageNames)
		logger.Debugf(dish)

		imageRowDb := model.Image{IdentRK: menuitem.Ident}

		logger.Debug("Проверка игнор-лист")
		for _, ignoreIdent := range cfg.RK7.MenuitemIdentIgnore {
			if menuitem.ItemIdent == ignoreIdent {
				logger.Warning("Блюдо в игнор-листе")
				imageRowDb.Status = sql.NullString{
					String: model.IMAGE_STATUS_IGNORE,
					Valid:  true,
				}
				err := imageRowDb.UpdateByIdentRKAndPos(db)
				if err != nil {
					return err
				}
				continue LoopOneStage
			}
		}

		if menuitem.Status != 3 {
			logger.Debug("Не активное блюдо. Пропускаем. Синхронизация меню отключила/удалила блюдо в WOO")
			menuitemsNotActive++
			imageRowDb.Status = sql.NullString{
				String: model.IMAGE_STATUS_IGNORE,
				Valid:  true,
			}
			err := imageRowDb.UpdateByIdentRKAndPos(db)
			if err != nil {
				return err
			}
		} else {
			logger.Debug("Блюдо активное. Проверяем цену/стоп-лист/включение синхронизации")
			menuitemsActive++

			logger.Debug("Проверяем наличие в стоп-листе")
			if dishRests, foundInStopList := dishRestsByIdent[menuitem.Ident]; foundInStopList {
				if dishRests.Prohibited == 1 || dishRests.Quantity == 0 {
					logger.Warning("Блюдо в стоп-листе")
					imageRowDb.Status = sql.NullString{
						String: model.IMAGE_STATUS_IGNORE,
						Valid:  true,
					}
					err := imageRowDb.UpdateByIdentRKAndPos(db)
					if err != nil {
						return err
					}
					continue LoopOneStage
				}
			}

			if menuitem.PRICETYPES == 9223372036854775807 {
				logger.Warning("Блюдо - цена не указана")
				imageRowDb.Status = sql.NullString{
					String: model.IMAGE_STATUS_IGNORE,
					Valid:  true,
				}
				err := imageRowDb.UpdateByIdentRKAndPos(db)
				if err != nil {
					return err
				}
				continue LoopOneStage
			}

			if menuitem.CLASSIFICATORGROUPS != cfg.RK7.CLASSIFICATORGROUPSALLOW {
				logger.Warning("Блюдо - с выключенной синхронизацией	")
				imageRowDb.Status = sql.NullString{
					String: model.IMAGE_STATUS_IGNORE,
					Valid:  true,
				}
				err := imageRowDb.UpdateByIdentRKAndPos(db)
				if err != nil {
					return err
				}
				continue LoopOneStage
			}
			if _, found := productsWooByID[menuitem.WOO_ID]; found {
				logger.Debug("Блюдо найдено в WOO")
				menuitemsFoundInWoo++
				logger.Debugf("Всего картинок %d штук", len(imageNames))
				for imageNameIndex, imageName := range imageNames { // todo отсюда все начинается
					if imageName == "" {
						logger.Warning("Не указано наименование картинки. Обнулить картинки в WOO")
						imageRowDb.Status = sql.NullString{
							String: model.IMAGE_STATUS_RK7_IMAGE_NAME_NOT_FOUND,
							Valid:  true,
						}
						err := imageRowDb.UpdateByIdentRKAndPos(db)
						if err != nil {
							return err
						}
					} else {
						logger.Debugf("Указано наименование картинки %s, позиция %d", imageName, imageNameIndex)
						logger.Debugf("Проверяем наличие картинки в папке %s", cfg.IMAGESYNC.Path)
						imageRowDb.Name = sql.NullString{
							String: imageName,
							Valid:  true,
						}
						imageRowDb.Pos = sql.NullInt32{
							Int32: int32(imageNameIndex),
							Valid: true,
						}
						path := fmt.Sprintf("%s/%s.jpg", cfg.IMAGESYNC.Path, imageName)
						if fileInfo, err := os.Stat(path); !os.IsNotExist(err) {
							logger.Debug("Получаем запись из DB")
							var imageInDb []model.Image
							query := "SELECT * FROM Image WHERE IdentRK = ? and Pos = ?"
							err = db.Select(&imageInDb, query, menuitem.ItemIdent, imageNameIndex)
							if err != nil {
								return errors.Wrapf(err, "failed SELECT to dbsqlite; dish %s; query %s(%d, %d)", dish, query, menuitem.ItemIdent, imageNameIndex)
							} else {
								switch {
								case len(imageInDb) == 0:
									logger.Warning("Запись не найдена в DB. Необходимо создать запись в DB и обновить картинку в WOO")
									imageRowDb.Status = sql.NullString{
										String: model.IMAGE_STATUS_NEED_UPDATE_BY_NOT_FOUND_IN_DB,
										Valid:  true,
									}
									err := imageRowDb.UpdateByIdentRKAndPos(db)
									if err != nil {
										return err
									}
								case len(imageInDb) == 1:
									logger.Debug("Блюдо найдено в DB. Сверяем имя/дату изменения")
									logger.Debugf("Наименование картинки: DB=%s, RK7=%s, File=%s", imageInDb[0].Name, imageName, fileInfo.Name())
									logger.Debugf("Дата изменения картинки: DB=%s, File=%s", imageInDb[0].ModTime.String, fileInfo.ModTime().Format(time.RFC3339))
									switch {
									case imageInDb[0].Name.String != imageName:
										logger.Debug("Наименование картинки изменилось. Обновляем картинку в WOO")
										imageRowDb.Status = sql.NullString{
											String: model.IMAGE_STATUS_NEED_UPDATE_BY_DIFF_NAME,
											Valid:  true,
										}
										err := imageRowDb.UpdateByIdentRKAndPos(db)
										if err != nil {
											return err
										}
									case imageInDb[0].ModTime.String != fileInfo.ModTime().Format(time.RFC3339):
										logger.Debug("Дата изменения картинки изменилась. Необходимо обновить картинку")
										imageRowDb.Status = sql.NullString{
											String: model.IMAGE_STATUS_NEED_UPDATE_BY_DIFF_DATE,
											Valid:  true,
										}
										err := imageRowDb.UpdateByIdentRKAndPos(db)
										if err != nil {
											return err
										}
									default:
										logger.Debug("Дата изменения картинки не изменилась. Наименование не изменилось. Необходимо проверить наличие в WOO")
										imageRowDb.Status = sql.NullString{
											String: model.IMAGE_STATUS_NO_NEED_UPDATE,
											Valid:  true,
										}
										err := imageRowDb.UpdateByIdentRKAndPos(db)
										if err != nil {
											return err
										}
									}
								case len(imageInDb) > 1:
									logger.Errorf("Недопустимое количество блюд в DB = %d > 1", len(imageInDb))
									imageRowDb.Status = sql.NullString{
										String: model.IMAGE_STATUS_NEED_UPDATE_BY_FIND_DOUBLE_IN_DB,
										Valid:  true,
									}
									err := imageRowDb.UpdateByIdentRKAndPos(db)
									if err != nil {
										return err
									}
								default:
									return errors.New("Неизвестная ошибка")
								}
							}
						} else {
							logger.Error("Файл картинки не найден. Отправить сообщение и обнулить картинку в WOO")
							imageRowDb.Status = sql.NullString{
								String: model.IMAGE_STATUS_FILE_NOT_FOUND,
								Valid:  true,
							}
							err := imageRowDb.UpdateByIdentRKAndPos(db)
							if err != nil {
								return err
							}
						}
					}
				} // todo тут все заканчивается
			} else {
				if menuitem.WOO_ID != 0 {
					logger.Error("Блюдо не найдено в WOO. Отправить сообщение")
					imageRowDb.Status = sql.NullString{
						String: model.IMAGE_STATUS_WOO_NOT_FOUND,
						Valid:  true,
					}
					err := imageRowDb.UpdateByIdentRKAndPos(db)
					if err != nil {
						return err
					}
				} else {
					logger.Warning("Блюдо без WOO_ID. Игнорировать")
					imageRowDb.Status = sql.NullString{
						String: model.IMAGE_STATUS_RK7_WOO_ID_NOT_FOUND,
						Valid:  true,
					}
					err := imageRowDb.UpdateByIdentRKAndPos(db)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	logger.Debug("Блюда WOO:")
	logger.Debugf("Всего: %d", len(productsWooByID))

	logger.Debug("Блюда RK7:")
	logger.Debugf("Всего: %d", len(menuitems))
	logger.Debugf("Активные: %d", menuitemsActive)
	logger.Debugf("Не активные: %d", menuitemsNotActive)

	os.Exit(3)
	//logger.Debugf("Не указана цена/синк выключен/в стоп-листе - сообщаем: %d", len(menuitemsPriceNotDefineStopListNoneSync)) //++
	//logger.Debugf("Игнорировано: %d", len(cfg.RK7.MenuitemIdentIgnore))
	//
	//logger.Debugf("Блюда найдено в WOO: %d", menuitemsFoundInWoo)                       //++
	//logger.Debugf("Блюдо не найдено в WOO - сообщить: %d", len(menuitemsNotFoundInWoo)) //++
	//logger.Debugf("Блюдо без WOO_ID - сообщить: %d", len(menuitemsWooIsNull))           //++
	//
	//logger.Debugf("Необходимо сообщить и обнулить картинку в WOO, причина - пустое поле WOO_IMAGE_NAME: %d", len(menuitemsImageNameNull)) //++
	//for _, item := range menuitemsImageNameNull {
	//	fmt.Println(item)
	//}
	//
	//logger.Debugf("Необходимо сообщить и обнулить картинку в WOO, причина - есть поле WOO_IMAGE_NAME, но нет картинки: %d", len(menuitemsImageFileNotFound)) //++
	//
	//logger.Debugf("Необходимо обновить картинку в WOO, причина - наименование изменилось: %d", len(menuitemsNeedUpdateInWooByName)) //++++
	//logger.Debugf("Необходимо обновить картинку в WOO, причина - дата изменилась: %d", len(menuitemsNeedUpdateInWooByDate))         //++++
	//
	//logger.Debugf("Необходимо проверить картинку в WOO, причина - дата не изменилась, наименование картинки не изменилось: %d", len(menuitemsNoNeedUpdateNeedVerifyInWoo)) //++
	//
	//logger.Debugf("Необходимо обновить картинку в WOO, причина - нет записи в DB: %d", len(menuitemsNeedAddInDB)) //++
	//logger.Debugf("Дубли в DB - сообщить: %d", len(menuitemsDubleInDB))                                           //++

	//os.Exit(322)
	/////////////////////////////////////////////////////////////////////////////
	//==========================================================================================///
	/////////////////////////////////////////////////////////////////////////////
	/*
		//++++++
		if len(menuitemsPriceNotDefineStopListNoneSync) > 0 {
			logger.Debugf("Найдены блюда без цены или выключен синк или в стоп-листе, отправляем сообщение")
			resultSyncAll = append(resultSyncAll, "<strong>Блюда RK7 без цены или выключен синк или в стоп-листе:</strong>")
			for _, menuitem := range menuitemsPriceNotDefineStopListNoneSync {
				dish := fmt.Sprintf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d, RK_Image: %s",
					menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES, menuitem.WOO_IMAGE_NAME)
				resultSyncAll = append(resultSyncAll, dish)
			}
		}
		//++++++
		if len(menuitemsNotFoundInWoo) > 0 {
			logger.Debugf("Блюда не найдены в WOO, отправляем сообщение")
			resultSyncAll = append(resultSyncAll, "<strong>Блюда RK7 не найдены в WOO, WOO_ID!=0:</strong>")
			resultSyncError = append(resultSyncError, "<strong>Блюда RK7 не найдены в WOO, WOO_ID!=0:</strong>")
			for _, menuitem := range menuitemsNotFoundInWoo {
				dish := fmt.Sprintf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d, RK_Image: %s",
					menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES, menuitem.WOO_IMAGE_NAME)
				resultSyncAll = append(resultSyncAll, dish)
				resultSyncError = append(resultSyncError, dish)
			}
		}
		//+
		if len(menuitemsWooIsNull) > 0 {
			logger.Debugf("Блюда RK7, WOO_ID не указан, отправляем сообщение")
			resultSyncAll = append(resultSyncAll, "<strong>Блюда RK7, WOO_ID не указан:</strong>")
			for _, menuitem := range menuitemsWooIsNull {
				dish := fmt.Sprintf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d, RK_Image: %s",
					menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES, menuitem.WOO_IMAGE_NAME)
				resultSyncAll = append(resultSyncAll, dish)
			}
		}

		//++++++
		if len(menuitemsImageNameNull) > 0 {
			logger.Debugf("Блюда RK7, WOO_IMAGE_NAME не указан, обнуляем в WOO и отправляем сообщение") // todo проверить обнуление
			resultSyncAll = append(resultSyncAll, "<strong>Блюда RK7, WOO_IMAGE_NAME не указан:</strong>")
			var mErrors []string
			mErrors = append(mErrors, "<strong>Блюда RK7, WOO_IMAGE_NAME не указан:</strong>")

			for i, menuitem := range menuitemsImageNameNull {
				dish := fmt.Sprintf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d, RK_Image: %s",
					menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES, menuitem.WOO_IMAGE_NAME)
				var messageText string
				err := NulledImageInWoo(menuitemsImageNameNull[i])
				if err != nil {
					messageText = fmt.Sprintf("%s; Не удалось обнулить в WOO: %v", dish, err)
					mErrors = append(mErrors, messageText)
					logger.Error(messageText)
				} else {
					messageText = fmt.Sprintf("%s; Успешно обнулено в WOO", dish)
					logger.Debug(messageText)
				}
				resultSyncAll = append(resultSyncAll, messageText)
			}
			if len(mErrors) > 1 {
				resultSyncError = append(resultSyncError, mErrors...)
			}
		}
		//++++++
		if len(menuitemsImageFileNotFound) > 0 {
			logger.Debugf("Блюда RK7, WOO_IMAGE_NAME указан, картинка не найдена в папке, обнуляем в WOO и отправляем сообщение")
			resultSyncAll = append(resultSyncAll, "<strong>Блюда RK7, картинки не найдены в папке:</strong>")
			resultSyncError = append(resultSyncError, "<strong>Блюда RK7, картинки не найдены в папке:</strong>")
			for i, menuitem := range menuitemsImageFileNotFound {
				dish := fmt.Sprintf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d, RK_Image: %s",
					menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES, menuitem.WOO_IMAGE_NAME)
				var messageText string
				err := NulledImageInWoo(menuitemsImageFileNotFound[i])
				if err != nil {
					messageText = fmt.Sprintf("%s; Не удалось обнулить в WOO: %v", dish, err)
					resultSyncError = append(resultSyncError, messageText)
					logger.Error(messageText)
				} else {
					messageText = fmt.Sprintf("%s; Успешно обнулено в WOO", dish)
					logger.Debug(messageText)
				}
				resultSyncAll = append(resultSyncAll, messageText)
			}
		}

		//++++++
		if len(menuitemsNeedUpdateInWooByName) > 0 {
			logger.Debugf("Блюда RK7, наименование картинки изменилась, обновляем/создаем картинку в WOO, кол-во %d", len(menuitemsNeedUpdateInWooByName))
			var messageText string
			var failedUpdateCount int
			var mErrors []string

			resultSyncAll = append(resultSyncAll, fmt.Sprintf("<strong>Блюда RK7, картинки изменились, обновляем в WOO, кол-во %d:</strong>", len(menuitemsNeedUpdateInWooByName)))
			mErrors = append(mErrors, fmt.Sprintf("<strong>Блюда RK7, картинки изменились, обновляем в WOO, кол-во %d:</strong>", len(menuitemsNeedUpdateInWooByName)))
			for i, menuitem := range menuitemsNeedUpdateInWooByName {
				dish := fmt.Sprintf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d, RK_Image: %s",
					menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES, menuitem.WOO_IMAGE_NAME)
				logger.Debug(dish)
				err := UpdateImageInWoo(menuitemsNeedUpdateInWooByName[i])
				if err != nil {
					messageText = fmt.Sprintf("%s; error: %v", dish, err)
					resultSyncAll = append(resultSyncAll, messageText)
					mErrors = append(mErrors, messageText)
					logger.Error(messageText)
					failedUpdateCount++
				} else {
					logger.Debug("Image успешно обновлен в WOO. Обновляем запись в DB")
					path := fmt.Sprintf("%s%s", cfg.IMAGESYNC.Path, menuitem.WOO_IMAGE_NAME)
					if fileInfo, err := os.Stat(path); !os.IsNotExist(err) {
						err := UpdateImageInDB(menuitemsNeedUpdateInWooByName[i], fileInfo.ModTime())
						if err != nil {
							messageText = fmt.Sprintf("%s; Не удалось обновить в DB: %v", dish, err)
							resultSyncAll = append(resultSyncAll, messageText)
							mErrors = append(mErrors, messageText)
							logger.Error(messageText)
							failedUpdateCount++
						} else {
							messageText = "Image успешно обновлен в DB"
							resultSyncAll = append(resultSyncAll, messageText)
							logger.Debug(messageText)
						}
					} else {
						messageText = fmt.Sprintf("%s; Не удалось найти картинку в папке при попытке добавить в DB", dish)
						resultSyncAll = append(resultSyncAll, messageText)
						mErrors = append(mErrors, messageText)
						logger.Error(messageText)
						failedUpdateCount++
					}
				}
			}

			if failedUpdateCount == 0 {
				messageText = fmt.Sprintf("Картинки успешно обновлены в WOO, кол-во %d", len(menuitemsNeedUpdateInWooByName))
				resultSyncAll = append(resultSyncAll, messageText)
			} else if failedUpdateCount < len(menuitemsNeedUpdateInWooByName) {
				messageText = fmt.Sprintf("Остальные картинки успешно обновлены в WOO, кол-во %d", len(menuitemsNeedUpdateInWooByName)-failedUpdateCount)
				resultSyncAll = append(resultSyncAll, messageText)
			}
			if len(mErrors) > 1 {
				resultSyncError = append(resultSyncError, mErrors...)
			}
		}
		//++++++
		if len(menuitemsNeedUpdateInWooByDate) > 0 {
			logger.Debugf("Блюда RK7, дата картинки изменилась, обновляем/создаем картинку в WOO, кол-во %d", len(menuitemsNeedUpdateInWooByDate))
			var messageText string
			var failedUpdateCount int
			var mErrors []string

			resultSyncAll = append(resultSyncAll, fmt.Sprintf("<strong>Блюда RK7, картинки изменились, обновляем в WOO, кол-во %d:</strong>", len(menuitemsNeedUpdateInWooByDate)))
			mErrors = append(mErrors, fmt.Sprintf("<strong>Блюда RK7, картинки изменились, обновляем в WOO, кол-во %d:</strong>", len(menuitemsNeedUpdateInWooByDate)))
			for i, menuitem := range menuitemsNeedUpdateInWooByDate {
				dish := fmt.Sprintf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d, RK_Image: %s",
					menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES, menuitem.WOO_IMAGE_NAME)
				logger.Debug(dish)
				err := UpdateImageInWoo(menuitemsNeedUpdateInWooByDate[i])
				if err != nil {
					messageText = fmt.Sprintf("%s; error: %v", dish, err)
					resultSyncAll = append(resultSyncAll, messageText)
					mErrors = append(mErrors, messageText)
					logger.Error(messageText)
					failedUpdateCount++
				} else {
					logger.Debug("Image успешно обновлен в WOO. Обновляем запись в DB")
					path := fmt.Sprintf("%s%s", cfg.IMAGESYNC.Path, menuitem.WOO_IMAGE_NAME)
					if fileInfo, err := os.Stat(path); !os.IsNotExist(err) {
						err := UpdateImageInDB(menuitemsNeedUpdateInWooByDate[i], fileInfo.ModTime())
						if err != nil {
							messageText = fmt.Sprintf("%s; Не удалось обновить в DB: %v", dish, err)
							resultSyncAll = append(resultSyncAll, messageText)
							mErrors = append(mErrors, messageText)
							logger.Error(messageText)
							failedUpdateCount++
						} else {
							messageText = "Image успешно обновлен в DB"
							resultSyncAll = append(resultSyncAll, messageText)
							logger.Debug(messageText)
						}
					} else {
						messageText = fmt.Sprintf("%s; Не удалось найти картинку в папке при попытке добавить в DB", dish)
						resultSyncAll = append(resultSyncAll, messageText)
						mErrors = append(mErrors, messageText)
						logger.Error(messageText)
						failedUpdateCount++
					}
				}
			}

			if failedUpdateCount == 0 {
				messageText = fmt.Sprintf("Картинки успешно обновлены в WOO, кол-во %d", len(menuitemsNeedUpdateInWooByDate))
				resultSyncAll = append(resultSyncAll, messageText)
			} else if failedUpdateCount < len(menuitemsNeedUpdateInWooByDate) {
				messageText = fmt.Sprintf("Остальные картинки успешно обновлены в WOO, кол-во %d", len(menuitemsNeedUpdateInWooByDate)-failedUpdateCount)
				resultSyncAll = append(resultSyncAll, messageText)
			}
			if len(mErrors) > 1 {
				resultSyncError = append(resultSyncError, mErrors...)
			}
		}

		//+++++-
		if len(menuitemsNoNeedUpdateNeedVerifyInWoo) > 0 {
			logger.Debugf("Блюда RK7, дата картинки не изменилась, наименование не изменилось, проверяем в WOO, что картинка существует, кол-во %d", len(menuitemsNoNeedUpdateNeedVerifyInWoo))

			var messageText string
			var failedUpdateCount int

			var mErrors []string
			mErrors = append(mErrors, fmt.Sprintf("<strong>Блюда RK7, дата картинок/наименование не изменилась, проверяем в WOO, кол-во %d:</strong>", len(menuitemsNoNeedUpdateNeedVerifyInWoo)))
			for i, menuitem := range menuitemsNoNeedUpdateNeedVerifyInWoo {
				dish := fmt.Sprintf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d, RK_Image: %s",
					menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES, menuitem.WOO_IMAGE_NAME)
				err := CheckImageInWoo(menuitemsNoNeedUpdateNeedVerifyInWoo[i])
				if err != nil {
					switch err.Error() {
					case ERROR_IMAGE_NOT_FOUND_IN_WOO:
						messageText = fmt.Sprintf("%s; %v", dish, err)
						mErrors = append(mErrors, messageText)
						failedUpdateCount++
					case ERROR_IMAGE_NOT_FOUND_IN_WOO_IMAGE:
						err = AddImageInWoo(menuitemsNoNeedUpdateNeedVerifyInWoo[i])
						if err != nil {
							messageText = fmt.Sprintf("%s; Не удалось создать в WOO: %v", dish, err)
							mErrors = append(mErrors, messageText)
							failedUpdateCount++
						} else {
							logger.Debug("Image успешно создан в Woo. Обновляем запись в DB")
							path := fmt.Sprintf("%s%s", cfg.IMAGESYNC.Path, menuitem.WOO_IMAGE_NAME)
							if fileInfo, err := os.Stat(path); !os.IsNotExist(err) {
								err := UpdateImageInDB(menuitemsNoNeedUpdateNeedVerifyInWoo[i], fileInfo.ModTime())
								if err != nil {
									messageText = fmt.Sprintf("%s; Не удалось обновить в DB: %v", dish, err)
									mErrors = append(mErrors, messageText)
									failedUpdateCount++
								} else {
									logger.Debug("Image успешно обновлен в DB")
									messageText = fmt.Sprintf("%s; Успешно обновлено: %v", dish, err)
								}
							} else {
								messageText = fmt.Sprintf("%s; Не удалось найти картинку в папке ", dish)
								mErrors = append(mErrors, messageText)
							}
						}
					case ERROR_IMAGE_CHECK_UNDEFINE:
						messageText = fmt.Sprintf("%s; %v", dish, err)
						mErrors = append(mErrors, messageText)
						failedUpdateCount++
					case ERROR_IMAGE_CHECK_ERROR_CAST:
						messageText = fmt.Sprintf("%s; %v", dish, err)
						mErrors = append(mErrors, messageText)
						failedUpdateCount++
					default:
						messageText = fmt.Sprintf("%s; %s", dish, "Неизвестная ошибка")
						mErrors = append(mErrors, messageText)
						failedUpdateCount++
					}
					resultSyncAll = append(resultSyncAll, messageText)
					logger.Debug(messageText)
				} else {
					messageText = "Картинка существует, обновление не требуется"
					logger.Debug(messageText)
					resultSyncAll = append(resultSyncAll, messageText)
				}
			}
			if len(mErrors) > 1 {
				resultSyncError = append(resultSyncError, mErrors...)
			}

			if failedUpdateCount == 0 {
				messageText = fmt.Sprintf("Картинки успешно проверены в WOO, кол-во %d", len(menuitemsNoNeedUpdateNeedVerifyInWoo))
			} else if failedUpdateCount < len(menuitemsNoNeedUpdateNeedVerifyInWoo) {
				messageText = fmt.Sprintf("Остальные картинки успешно проверены в WOO, кол-во %d", len(menuitemsNoNeedUpdateNeedVerifyInWoo)-failedUpdateCount)
			} else if failedUpdateCount == len(menuitemsNoNeedUpdateNeedVerifyInWoo) {
				messageText = fmt.Sprintf("Ни одна картинка не проверена в WOO, кол-во %d", len(menuitemsNoNeedUpdateNeedVerifyInWoo)-failedUpdateCount)
			} else {
				messageText = "Неизвестная ошибка при проверке картинки в WOO"
			}
			logger.Debug(messageText)
			resultSyncAll = append(resultSyncAll, messageText)
		}

		//++++++
		if len(menuitemsNeedAddInDB) > 0 {
			logger.Debugf("Блюда RK7, нет записи в БД, обновляем/создаем картинку в WOO и добавляем запись в DB, кол-во %d", len(menuitemsNeedAddInDB))
			resultSyncAll = append(resultSyncAll, fmt.Sprintf("<strong>Блюда RK7, нет записи в DB, обновляем в WOO, кол-во %d:</strong>", len(menuitemsNeedAddInDB)))
			var messageText string
			var failedUpdateCount int

			var mErrors []string
			mErrors = append(mErrors, fmt.Sprintf("<strong>Блюда RK7, нет записи в DB, обновляем в WOO, кол-во %d:</strong>", len(menuitemsNeedAddInDB)))
			for i, menuitem := range menuitemsNeedAddInDB {
				dish := fmt.Sprintf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d, RK_Image: %s",
					menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES, menuitem.WOO_IMAGE_NAME)
				logger.Debug(dish)
				err := UpdateImageInWoo(menuitemsNeedAddInDB[i])
				if err != nil {
					messageText = fmt.Sprintf("%s; error: %v", dish, err)
					resultSyncAll = append(resultSyncAll, messageText)
					mErrors = append(mErrors, messageText)
					logger.Error(messageText)
					failedUpdateCount++
				} else {
					logger.Debug("Image успешно обновлен в WOO. Обновляем запись в DB")
					path := fmt.Sprintf("%s%s", cfg.IMAGESYNC.Path, menuitem.WOO_IMAGE_NAME)
					if fileInfo, err := os.Stat(path); !os.IsNotExist(err) {
						err := UpdateImageInDB(menuitemsNeedAddInDB[i], fileInfo.ModTime())
						if err != nil {
							messageText = fmt.Sprintf("%s; Не удалось обновить в DB: %v", dish, err)
							resultSyncAll = append(resultSyncAll, messageText)
							mErrors = append(mErrors, messageText)
							logger.Error(messageText)
							failedUpdateCount++
						} else {
							messageText = "Image успешно обновлен в DB"
							resultSyncAll = append(resultSyncAll, messageText)
							logger.Debug(messageText)
						}
					} else {
						messageText = fmt.Sprintf("%s; Не удалось найти картинку в папке при попытке добавить в DB", dish)
						resultSyncAll = append(resultSyncAll, messageText)
						mErrors = append(mErrors, messageText)
						logger.Error(messageText)
						failedUpdateCount++
					}
				}
			}

			if failedUpdateCount == 0 {
				messageText = fmt.Sprintf("Картинки успешно обновлены в WOO, кол-во %d", len(menuitemsNeedAddInDB))
				resultSyncAll = append(resultSyncAll, messageText)
			} else if failedUpdateCount < len(menuitemsNeedAddInDB) {
				messageText = fmt.Sprintf("Остальные картинки успешно обновлены в WOO, кол-во %d", len(menuitemsNeedAddInDB)-failedUpdateCount)
				resultSyncAll = append(resultSyncAll, messageText)
			}
			if len(mErrors) > 1 {
				resultSyncError = append(resultSyncError, mErrors...)
			}
		}

		//++++++
		if len(menuitemsDubleInDB) > 0 {
			logger.Debugf("Блюда RK7, имеются дубли в DB, отправляем сообщение")
			resultSyncAll = append(resultSyncAll, "<strong>Блюда RK7, имеются дубли в DB, некорректная ситуация:</strong>")
			resultSyncError = append(resultSyncError, "<strong>Блюда RK7, имеются дубли в DB, некорректная ситуация:</strong>")
			for _, menuitem := range menuitemsDubleInDB {
				dish := fmt.Sprintf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d, RK_Image: %s",
					menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES, menuitem.WOO_IMAGE_NAME)
				resultSyncAll = append(resultSyncAll, dish)
				resultSyncError = append(resultSyncError, dish)
			}
		}

		//++++++
		if len(menuitemsNonJpgFile) > 0 {
			logger.Debugf("Картинки формата не jpg")
			resultSyncAll = append(resultSyncAll, "<strong>Картинки формата не jpg:</strong>")
			resultSyncError = append(resultSyncError, "<strong>Картинки формата не jpg:</strong>")
			for _, menuitem := range menuitemsNonJpgFile {
				dish := fmt.Sprintf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d, RK_Image: %s",
					menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES, menuitem.WOO_IMAGE_NAME)
				resultSyncAll = append(resultSyncAll, dish)
				resultSyncError = append(resultSyncError, dish)
			}
		}

		if len(resultSyncError) > 0 {
			logger.Debug("Cинхронизация картинок завершилась с ошибками")
			if cfg.MENUSYNC.TelegramReport == 1 {
				telegram.SendMessageToTelegramWithLogError(strings.Join(resultSyncAll, "\n"))
			} else if cfg.MENUSYNC.TelegramReport == 2 {
				telegram.SendMessageToTelegramWithLogError(strings.Join(resultSyncError, "\n"))
			}
		} else {
			logger.Debug("Cинхронизация картинок завершилась успешно")
			if cfg.MENUSYNC.TelegramReport == 1 {
				telegram.SendMessageToTelegramWithLogError("Cинхронизация картинок завершилась успешно")
			}
		}

	*/
	/////////////////////////////////////////////////////////////////////////////
	//==========================================================================================///
	/////////////////////////////////////////////////////////////////////////////

	return nil
}

//func Handler

/*
func NulledImageInWoo(menuitem *modelsRK7API.MenuitemItem) error {
	logger := logging.GetLogger()
	logger.Debug("Start NulledImageInWoo")
	defer logger.Debug("End NulledImageInWoo")

	menu, err := cache.GetMenu()
	if err != nil {
		return err
	}

	productsWooByID, err := menu.GetProductsWooByID()
	if err != nil {
		return err
	}

	cfg := config.GetConfig()

	if _, found := productsWooByID[menuitem.WOO_ID]; found {
		if len(productsWooByID[menuitem.WOO_ID].Images) == 0 {
			logger.Debug("Удаление картинок не требуется")
			return nil
		} else {
			for _, image := range productsWooByID[menuitem.WOO_ID].Images {
				logger.Debugf("Приступаем к удалению файла в WOO ID=%d", image.Id)
				status, _ := SendDeleteRequest(fmt.Sprintf("%s/wp-json/wp/v2/media/%d", cfg.WOOCOMMERCE.URL, image.Id))
				switch status {
				case "200 OK":
					logger.Debug("Картинка удалена!")
					logger.Debug("Обнуляем запись в DB")
					err := NulledImageInDB(menuitem)
					if err != nil {
						return errors.Wrap(err, "Ошибка при обнулении записи в DB")
					} else {
						logger.Debug("Обнулена запись в DB")
						return nil
					}
				default:
					return errors.New(fmt.Sprintf("Ошибка при попытке удалить запись из DB; status: %s", status))
				}
			}
		}
	} else {
		return errors.New("Продукт не найден")
	}

	return nil
}

func NulledImageInDB(menuitem *modelsRK7API.MenuitemItem) error {
	logger := logging.GetLogger()
	logger.Debug("Start NulledImageInDB")
	defer logger.Debug("End NulledImageInDB")

	db, err := sqlx.Connect("sqlite3", database.DB_NAME)
	if err != nil {
		logger.Fatalf("failed sqlx.Connect; %v", err)
	}
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			logger.Fatalf("failed close sqlx.Connect, err: %v", err)
		}
	}(db)

	tx := db.MustBegin()
	tx.MustExec("DELETE FROM Image WHERE IdentRK = $1;", menuitem.ItemIdent)
	err = tx.Commit()
	if err != nil {
		return errors.Wrapf(err, "failed DELETE in dbsqlite")
	} else {
		return nil
	}
}

func UpdateImageInWoo(menuitem *modelsRK7API.MenuitemItem) error {

	logger := logging.GetLogger()
	logger.Debug("Start UpdateImageInWoo")
	defer logger.Debug("End UpdateImageInWoo")

	menu, err := cache.GetMenu()
	if err != nil {
		return err
	}

	productsWooByID, err := menu.GetProductsWooByID()
	if err != nil {
		return err
	}

	if product, found := productsWooByID[menuitem.WOO_ID]; found {
		if len(product.Images) == 0 {
			logger.Debug("Картинок нет в WOO. Добавляем")
			err := AddImageInWoo(menuitem)
			if err != nil {
				logger.Error("UpdateImageInWoo:Картинка не добавлена в продукт в WOO")
				return errors.Wrap(err, "UpdateImageInWoo:Картинка не добавлена в продукт в WOO")
			} else {
				logger.Debug("Картинка успешно добавлена в продукт в WOO")
				return nil
			}
		} else {
			var findImages []models.ProductImage // картинку будут обновлены
			var delImages []models.ProductImage  // картинки будут удалены

			for _, image := range productsWooByID[menuitem.WOO_ID].Images {
				logger.Debugf("Картинка %v", image)
				if menuitem.WOO_IMAGE_NAME != "" {
					logger.Debug("Выполняем поиск rkeeper картинки в WOO")
					if image.Alt == menuitem.WOO_IMAGE_NAME {
						logger.Debug("Картинка WOO совпадает с rkeeper")
						findImages = append(findImages, image)
					} else {
						logger.Debug("Картинка WOO не совпадает с rkeeper. Будет удалена позже")
						delImages = append(delImages, image)
					}
				} else {
					logger.Error("menuitem.WOO_IMAGE_NAME не указан")
					return errors.New("menuitem.WOO_IMAGE_NAME не указан")
				}
			}

			cfg := config.GetConfig()
			logger.Debug("Обнуляем WOO:IMAGES")
			if len(findImages) > 0 {
				logger.Debugf("Будет удалено %d картинок", len(delImages))
				for _, image := range findImages {
					logger.Debugf("Приступаем к удалению картинки в WOO ID=%d", image.Id)
					status, _ := SendDeleteRequest(fmt.Sprintf("%s/wp-json/wp/v2/media/%d", cfg.WOOCOMMERCE.URL, image.Id))
					switch status {
					case "200 OK":
						logger.Debug("Картинка удалена!")
					default:
						return errors.New(fmt.Sprintf("Неизвестная ошибка при удалении картинки, image.id=%d", image.Id))
					}
				}
			}
			if len(delImages) > 0 {
				logger.Debugf("Будет удалено %d картинок", len(delImages))
				for _, image := range delImages {
					logger.Debugf("Приступаем к удалению картинки в WOO ID=%d", image.Id)
					status, _ := SendDeleteRequest(fmt.Sprintf("%s/wp-json/wp/v2/media/%d", cfg.WOOCOMMERCE.URL, image.Id))
					switch status {
					case "200 OK":
						logger.Debug("Картинка удалена!")
					default:
						return errors.New(fmt.Sprintf("Неизвестная ошибка при удалении картинки, image.id=%d", image.Id))
					}
				}
			}

			productsWooByID[menuitem.WOO_ID].Images = make([]models.ProductImage, 0)
			logger.Debug("WOO:IMAGES обнулена")
			logger.Debug("Приступаем к закачке картинки")

			//api := wooapi.GetAPI()
			//product, err := api.ProductGet(productsWooByID[menuitem.WOO_ID].ID)
			//if err != nil {
			//	return err
			//}
			//logger.Error(product.Images)
			//
			//os.Exit(2)
			err = AddImageInWoo(menuitem)
			if err != nil {
				logger.Error("UpdateImageInWoo:AddImageInWoo:Картинка не добавлена в продукт в WOO")
				return errors.Wrap(err, "UpdateImageInWoo:AddImageInWoo:Картинка не добавлена в продукт в WOO")
			} else {
				logger.Debug("Картинка успешно добавлена в продукт в WOO")
				return nil
			}
			/////////////////////////////////////////////////////////////////////////////
			//==========================================================================================///
			/////////////////////////////////////////////////////////////////////////////
			//if len(findImages) == 0 {
			//	logger.Debug("Приступаем к закачке картинки")
			//	err := AddImageInWoo(menuitem)
			//	if err != nil {
			//		logger.Error("UpdateImageInWoo:AddImageInWoo:Картинка не добавлена в продукт в WOO")
			//		return errors.Wrap(err, "UpdateImageInWoo:AddImageInWoo:Картинка не добавлена в продукт в WOO")
			//	} else {
			//		logger.Debug("Картинка успешно добавлена в продукт в WOO")
			//	}
			//} else {
			//	logger.Debugf("Картинок найдено %d штук. Обнуляем все и повторно закачиваем", len(findImages))
			//	for _, image := range findImages {
			//		logger.Debugf("Картинка id=%d", image.Id)
			//		status, _ := SendDeleteRequest(fmt.Sprintf("%s/wp-json/wp/v2/media/%d", cfg.WOOCOMMERCE.URL, image.Id))
			//		switch status {
			//		case "200 OK":
			//			logger.Debug("Картинка удалена!")
			//		default:
			//			errorText := fmt.Sprintf("Неизвестная ошибка при удалении, id=%d", image.Id)
			//			logger.Debug(errorText)
			//			return errors.New(errorText)
			//		}
			//	}
			//	logger.Debug("Приступаем к закачке картинок")
			//	err := AddImageInWoo(menuitem)
			//	if err != nil {
			//		logger.Error("UpdateImageInWoo:AddImageInWoo:Картинка не добавлена в продукт в WOO")
			//		return errors.Wrap(err, "UpdateImageInWoo:AddImageInWoo:Картинка не добавлена в продукт в WOO")
			//	} else {
			//		logger.Debug("Картинка успешно добавлена в продукт в WOO")
			//	}
			//}
			//
			//if len(delImages) > 0 {
			//	logger.Debugf("Будет удалено %d картинок", len(delImages))
			//	for _, image := range delImages {
			//		logger.Debugf("Приступаем к удалению картинки в WOO ID=%d", image.Id)
			//		status, _ := SendDeleteRequest(fmt.Sprintf("%s/wp-json/wp/v2/media/%d", cfg.WOOCOMMERCE.URL, image.Id))
			//		switch status {
			//		case "200 OK":
			//			logger.Debug("Картинка удалена!")
			//		default:
			//			return errors.New(fmt.Sprintf("Неизвестная ошибка при удалении картинки, image.id=%d", image.Id))
			//		}
			//	}
			//}
			//return nil

		}
	} else {
		return errors.New("Продукт не найден в WOO")
	}
}

func AddImageInWoo(menuitem *modelsRK7API.MenuitemItem) error {
	logger := logging.GetLogger()
	logger.Debug("Start AddImageInWoo")
	defer logger.Debug("End AddImageInWoo")
	menu, err := cache.GetMenu()
	if err != nil {
		return err
	}
	productsWooByID, err := menu.GetProductsWooByID()
	if err != nil {
		return err
	}
	cfg := config.GetConfig()
	api := wooapi.GetAPI()
	status, content := SendPostRequest(fmt.Sprintf("%s/wp-json/wp/v2/media", cfg.WOOCOMMERCE.URL), menuitem.WOO_IMAGE_NAME)
	switch status {
	case "201 Created":
		logger.Debug("Картинка закачана успешно")
		logger.Debug(status)
		logger.Debug(string(content))

		imageJ := new(imageJson)
		err := json.Unmarshal(content, imageJ)
		if err != nil {
			return errors.Wrap(err, "Не удалось выполнить Unmarshal")
		} else {
			if _, found := productsWooByID[menuitem.WOO_ID]; found {
				var imageAdd models.ProductImage
				imageAdd.Alt = menuitem.WOO_IMAGE_NAME
				imageAdd.Id = imageJ.Id
				logger.Debugf("Добавляем картинку: ID:%d, Name:%s", imageAdd.Id, imageAdd.Alt)
				productsWooByID[menuitem.WOO_ID].Images = append(productsWooByID[menuitem.WOO_ID].Images, imageAdd)
				_, err := api.ProductUpdate(productsWooByID[menuitem.WOO_ID])
				if err != nil {
					logger.Error(err, "Картинка не обновилась у продукт")
					return errors.Wrapf(err, "Картинка не обновилась у продукт")
				} else {
					logger.Debug("Картинка обновлена")
					return nil
				}
			} else {
				return errors.New("Продукт не найден")
			}
		}
	default:
		errorText := fmt.Sprintf("Неизвестное поведение. Status: %s", status)
		logger.Error(errorText)
		return errors.New(errorText)
	}
}

func SendPostRequest(url string, filename string) (string, []byte) {
	client := &http.Client{}

	cfg := config.GetConfig()
	path := fmt.Sprintf("%s%s", cfg.IMAGESYNC.Path, filename)
	data, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	req, err := http.NewRequest("POST", url, data)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "image/jpeg")
	req.Header.Set("Content-Disposition", fmt.Sprintf(`form-data; filename="%s"`, filename))

	req.SetBasicAuth(cfg.WOOCOMMERCE.User, cfg.WOOCOMMERCE.Password)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return resp.Status, content
}

func SendDeleteRequest(url string) (string, []byte) {

	cfg := config.GetConfig()

	client := &http.Client{}

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	q := req.URL.Query()
	q.Add("force", "true")
	req.URL.RawQuery = q.Encode()
	req.SetBasicAuth(cfg.WOOCOMMERCE.User, cfg.WOOCOMMERCE.Password)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return resp.Status, content
}

func CheckImageInWoo(menuitem *modelsRK7API.MenuitemItem) error {
	//todo срань
	logger := logging.GetLogger()
	logger.Debug("Start CheckImageInWoo")
	defer logger.Debug("End CheckImageInWoo")

	menu, err := cache.GetMenu()
	if err != nil {
		return err
	}

	productsWooByID, err := menu.GetProductsWooByID()
	if err != nil {
		return err
	}

	if product, found := productsWooByID[menuitem.WOO_ID]; found {
		if len(product.Images) > 0 {
			logger.Debug("Картинок более 0 в WOO:Images")
			logger.Debug("Выполняем поиск картинки в WOO:Images")
			for _, image := range product.Images {
				if image.Alt == menuitem.WOO_IMAGE_NAME && menuitem.WOO_IMAGE_NAME != "" {
					logger.Debug("Картинка найдена")
					return nil
				}
			}
			logger.Debug("Не удалось найти картинку в WOO:Image. Необходимо обновить картинку")
			return errors.New(ERROR_IMAGE_NOT_FOUND_IN_WOO_IMAGE)
		} else {
			logger.Debug("Картинок нет в WOO:Images")
			return errors.New(ERROR_IMAGE_NOT_FOUND_IN_WOO_IMAGE)
		}
	} else {
		logger.Debug("Не удалось найти product в WOO")
		return errors.New(ERROR_IMAGE_NOT_FOUND_IN_WOO)
	}
}

func UpdateImageInDB(menuitem *modelsRK7API.MenuitemItem, modTime time.Time) error {

	logger := logging.GetLogger()
	logger.Debug("Start UpdateImageInDB")
	defer logger.Debug("End UpdateImageInDB")

	db, err := sqlx.Connect("sqlite3", database.DB_NAME)
	if err != nil {
		logger.Fatalf("failed sqlx.Connect; %v", err)
	}
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			logger.Fatalf("failed close sqlx.Connect, err: %v", err)
		}
	}(db)

	var menuitemDBs []database.Image
	query := fmt.Sprintf(database.DATABASE_SELECT_MENUITEM, menuitem.ItemIdent)
	err = db.Select(&menuitemDBs, query)
	if err != nil {
		return errors.Wrapf(err, "failed SELECT to dbsqlite")
	} else {
		if len(menuitemDBs) == 0 {
			tx := db.MustBegin()
			tx.MustExec("INSERT INTO Image (IdentRK, ImageModTime, ImageName) VALUES ($1, $2, $3)", menuitem.ItemIdent, modTime.Format(time.RFC3339), menuitem.WOO_IMAGE_NAME)
			err := tx.Commit()
			if err != nil {
				return errors.Wrapf(err, "failed INSERT to dbsqlite")
			} else {
				return nil
			}
		} else if len(menuitemDBs) == 1 {
			tx := db.MustBegin()
			tx.MustExec("UPDATE Image SET ImageModTime = $1 WHERE IdentRK = $2", modTime.Format(time.RFC3339), menuitem.ItemIdent)
			tx.MustExec("UPDATE Image SET ImageName = $1 WHERE IdentRK = $2", menuitem.WOO_IMAGE_NAME, menuitem.ItemIdent)
			err := tx.Commit()
			if err != nil {
				return errors.Wrapf(err, "failed UPDATE to dbsqlite")
			} else {
				return nil
			}
		} else if len(menuitemDBs) >= 1 {
			return errors.Wrapf(err, "result select over 1")
		} else {
			return errors.Wrapf(err, "Неизвестная ошибка")
		}
	}
}

const (
	ERROR_IMAGE_NOT_FOUND_IN_WOO       = "Не найдена картинка в WOO"
	ERROR_IMAGE_NOT_FOUND_IN_WOO_IMAGE = "Не найдена картинка в WOO:Image"
	ERROR_IMAGE_CHECK_ERROR_CAST       = "Ошибка приведения"
	ERROR_IMAGE_CHECK_UNDEFINE         = "Неизвестная ошибка"
)

type imageJson struct {
	Id      int    `json:"id"`
	Date    string `json:"date"`
	DateGmt string `json:"date_gmt"`
	Guid    struct {
		Rendered string `json:"rendered"`
		Raw      string `json:"raw"`
	} `json:"guid"`
	Modified    string `json:"modified"`
	ModifiedGmt string `json:"modified_gmt"`
	Slug        string `json:"slug"`
	Status      string `json:"status"`
	Type        string `json:"type"`
	Link        string `json:"link"`
	Title       struct {
		Raw      string `json:"raw"`
		Rendered string `json:"rendered"`
	} `json:"title"`
	Author            int           `json:"author"`
	CommentStatus     string        `json:"comment_status"`
	PingStatus        string        `json:"ping_status"`
	Template          string        `json:"template"`
	Meta              []interface{} `json:"meta"`
	PermalinkTemplate string        `json:"permalink_template"`
	GeneratedSlug     string        `json:"generated_slug"`
	Acf               []interface{} `json:"acf"`
	Description       struct {
		Raw      string `json:"raw"`
		Rendered string `json:"rendered"`
	} `json:"description"`
	Caption struct {
		Raw      string `json:"raw"`
		Rendered string `json:"rendered"`
	} `json:"caption"`
	AltText      string `json:"alt_text"`
	MediaType    string `json:"media_type"`
	MimeType     string `json:"mime_type"`
	MediaDetails struct {
		Width    int    `json:"width"`
		Height   int    `json:"height"`
		File     string `json:"file"`
		Filesize int    `json:"filesize"`
		Sizes    struct {
			Medium struct {
				File      string `json:"file"`
				Width     int    `json:"width"`
				Height    int    `json:"height"`
				Filesize  int    `json:"filesize"`
				MimeType  string `json:"mime_type"`
				SourceUrl string `json:"source_url"`
			} `json:"medium"`
			Thumbnail struct {
				File      string `json:"file"`
				Width     int    `json:"width"`
				Height    int    `json:"height"`
				Filesize  int    `json:"filesize"`
				MimeType  string `json:"mime_type"`
				SourceUrl string `json:"source_url"`
			} `json:"thumbnail"`
			WoocommerceThumbnail struct {
				File      string `json:"file"`
				Width     int    `json:"width"`
				Height    int    `json:"height"`
				Filesize  int    `json:"filesize"`
				Uncropped bool   `json:"uncropped"`
				MimeType  string `json:"mime_type"`
				SourceUrl string `json:"source_url"`
			} `json:"woocommerce_thumbnail"`
			WoocommerceSingle struct {
				File      string `json:"file"`
				Width     int    `json:"width"`
				Height    int    `json:"height"`
				Filesize  int    `json:"filesize"`
				MimeType  string `json:"mime_type"`
				SourceUrl string `json:"source_url"`
			} `json:"woocommerce_single"`
			WoocommerceGalleryThumbnail struct {
				File      string `json:"file"`
				Width     int    `json:"width"`
				Height    int    `json:"height"`
				Filesize  int    `json:"filesize"`
				MimeType  string `json:"mime_type"`
				SourceUrl string `json:"source_url"`
			} `json:"woocommerce_gallery_thumbnail"`
			QuickViewImageSize struct {
				File      string `json:"file"`
				Width     int    `json:"width"`
				Height    int    `json:"height"`
				Filesize  int    `json:"filesize"`
				MimeType  string `json:"mime_type"`
				SourceUrl string `json:"source_url"`
			} `json:"quick_view_image_size"`
			Full struct {
				File      string `json:"file"`
				Width     int    `json:"width"`
				Height    int    `json:"height"`
				MimeType  string `json:"mime_type"`
				SourceUrl string `json:"source_url"`
			} `json:"full"`
		} `json:"sizes"`
		ImageMeta struct {
			Aperture         string        `json:"aperture"`
			Credit           string        `json:"credit"`
			Camera           string        `json:"camera"`
			Caption          string        `json:"caption"`
			CreatedTimestamp string        `json:"created_timestamp"`
			Copyright        string        `json:"copyright"`
			FocalLength      string        `json:"focal_length"`
			Iso              string        `json:"iso"`
			ShutterSpeed     string        `json:"shutter_speed"`
			Title            string        `json:"title"`
			Orientation      string        `json:"orientation"`
			Keywords         []interface{} `json:"keywords"`
		} `json:"image_meta"`
	} `json:"media_details"`
	Post              interface{} `json:"post"`
	SourceUrl         string      `json:"source_url"`
	MissingImageSizes []string    `json:"missing_image_sizes"`
	Links             struct {
		Self []struct {
			Href string `json:"href"`
		} `json:"self"`
		Collection []struct {
			Href string `json:"href"`
		} `json:"collection"`
		About []struct {
			Href string `json:"href"`
		} `json:"about"`
		Author []struct {
			Embeddable bool   `json:"embeddable"`
			Href       string `json:"href"`
		} `json:"author"`
		Replies []struct {
			Embeddable bool   `json:"embeddable"`
			Href       string `json:"href"`
		} `json:"replies"`
		WpActionUnfilteredHtml []struct {
			Href string `json:"href"`
		} `json:"wp:action-unfiltered-html"`
		WpActionAssignAuthor []struct {
			Href string `json:"href"`
		} `json:"wp:action-assign-author"`
		Curies []struct {
			Name      string `json:"name"`
			Href      string `json:"href"`
			Templated bool   `json:"templated"`
		} `json:"curies"`
	} `json:"_links"`
}
*/
