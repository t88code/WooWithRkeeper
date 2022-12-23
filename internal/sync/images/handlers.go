package images

import (
	"WooWithRkeeper/internal/cache"
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/internal/database/model/image"
	"WooWithRkeeper/internal/database/model/imagefile"
	"WooWithRkeeper/internal/rk7api/models"
	"WooWithRkeeper/internal/telegram"
	"WooWithRkeeper/internal/wooapi"
	models2 "WooWithRkeeper/internal/wooapi/models"
	"WooWithRkeeper/pkg/logging"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
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

// HandlerImageFileToDb 1 этап - Синхронизация файлов картинок из папки с Woo.Media/DB.ImageFiles
func HandlerImageFileToDb(db *sqlx.DB) error {

	logger := logging.GetLogger()
	logger.Debug("Start HandlerImageFileToDb")
	defer logger.Debug("End HandlerImageFileToDb")
	var err error
	cfg := config.GetConfig()

	files, err := ioutil.ReadDir(cfg.IMAGESYNC.Path)
	if err != nil {
		return errors.Wrap(err, "failed in ioutil.ReadDir")
	}

	for _, file := range files {
		match, _ := regexp.MatchString(".jpg$", file.Name())
		if match {
			logger.Debug("Поиск картинки в DB")
			var imageFilesInDb []imagefile.ImageFile
			query := "SELECT * FROM ImageFile WHERE Name = ?"
			err = db.Select(&imageFilesInDb, query, file.Name())
			if err != nil {
				return errors.Wrapf(err, "failed SELECT to dbsqlite; query %s(%s)", query, file.Name())
			} else {
				switch {
				case len(imageFilesInDb) == 0:
					// 1 - добавляем запись в DB
					// 2 - закачиваем картинку
					// 3 - обновляем запись в DB - ID-WOO
					logger.Debug("Запись не найдена в DB. Необходимо создать запись в DB и закачать картинку в WOO")
					err := AddImageToDb(db, file)
					if err != nil {
						return errors.Wrap(err, "failed in AddImageToDb")
					} else {
						logger.Debug("Строка добавлена успешно")
						err = UploadImageAndUpdateDb(db, file)
						if err != nil {
							return errors.Wrap(err, "failed in UploadImageAndUpdateDb")
						}
					}
				case len(imageFilesInDb) == 1:
					// 1 - сверяем modTime
					// если отличается, то
					// - удаляем картинку из WOO
					// - закачиваем картинку
					// - обновляем запись в DB - ID-WOO
					// если не отличается, то
					// - проверяем что картинка существует
					// - если не существует то закачиваем картинку
					// - обновляем запись в DB - ID-WOO
					logger.Debug("Запись найдена в DB")
					if imageFilesInDb[0].ModTime.Valid {
						if imageFilesInDb[0].ModTime.String != file.ModTime().Format(time.RFC3339) {
							logger.Debugf("Дата картинки не совпадет; DB=%s, File=%s", imageFilesInDb[0].ModTime.String, file.ModTime().Format(time.RFC3339))
							if imageFilesInDb[0].IdentWOO.Valid {
								logger.Debugf("Идентификатор(%d) найден - удаляем в Woo", imageFilesInDb[0].IdentWOO.Int32)
								err = DeleteImageInWoo(int(imageFilesInDb[0].IdentWOO.Int32))
								if err != nil {
									return errors.Wrap(err, "failed in DeleteImageInWoo")
								}
							}
							err = UploadImageAndUpdateDb(db, file)
							if err != nil {
								return errors.Wrap(err, "failed in UploadImageAndUpdateDb")
							}
						} else {
							logger.Debug("Дата картинки совпадет")
							if imageFilesInDb[0].IdentWOO.Valid {
								id, err := CheckImageInWoo(int(imageFilesInDb[0].IdentWOO.Int32))
								if err != nil {
									return errors.Wrap(err, "failed in CheckImageInWoo")
								}
								if id == 0 {
									err = UploadImageAndUpdateDb(db, file)
									if err != nil {
										return errors.Wrap(err, "failed in UploadImageAndUpdateDb")
									}
								}
							} else {
								return errors.New(fmt.Sprintf("ID картинки в базе не указан для картинки: %v", imageFilesInDb[0]))
							}
						}
					}
				case len(imageFilesInDb) > 1:
					return errors.New(fmt.Sprintf("Недопустимое количество блюд в DB = %d > 1", len(imageFilesInDb)))
				default:
					return errors.New("Неизвестная ошибка")
				}
			}
		}
	}
	return nil
}

// HandlerMenuitemsToDbImage 2 этап - Синхронизация Menuitems c DB.Image
func HandlerMenuitemsToDbImage(db *sqlx.DB) error {

	logger := logging.GetLogger()
	logger.Debug("Start HandlerMenuitemsToDbImage")
	defer logger.Debug("End HandlerMenuitemsToDbImage")
	var err error
	cfg := config.GetConfig()

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

		imageRowDb := image.Image{IdentRK: menuitem.Ident}

		logger.Debug("Проверка игнор-лист")
		for _, ignoreIdent := range cfg.RK7.MenuitemIdentIgnore {
			if menuitem.ItemIdent == ignoreIdent {
				logger.Debug("Блюдо в игнор-листе")
				imageRowDb.Status = sql.NullString{
					String: image.IMAGE_STATUS_IGNORE,
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
				String: image.IMAGE_STATUS_IGNORE,
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
					logger.Debug("Блюдо в стоп-листе")
					imageRowDb.Status = sql.NullString{
						String: image.IMAGE_STATUS_IGNORE,
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
				logger.Debug("Блюдо - цена не указана")
				imageRowDb.Status = sql.NullString{
					String: image.IMAGE_STATUS_IGNORE,
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
					imageRowDb := image.Image{IdentRK: menuitem.Ident}
					imageRowDb.Pos = sql.NullInt32{
						Int32: int32(imageNameIndex),
						Valid: true,
					}
					if imageName == "" {
						logger.Debug("Не указано наименование картинки. Обнулить картинки в WOO")
						imageRowDb.Status = sql.NullString{
							String: image.IMAGE_STATUS_RK7_IMAGE_NAME_NOT_FOUND,
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
						path := fmt.Sprintf("%s/%s", cfg.IMAGESYNC.Path, imageName)
						if fileInfo, err := os.Stat(path); !os.IsNotExist(err) {
							logger.Debug("Получаем запись из DB")
							var imageInDb []image.Image
							query := "SELECT * FROM Image WHERE IdentRK = ? and Pos = ?"
							err = db.Select(&imageInDb, query, menuitem.ItemIdent, imageNameIndex)
							if err != nil {
								return errors.Wrapf(err, "failed SELECT to dbsqlite; dish %s; query %s(%d, %d)", dish, query, menuitem.ItemIdent, imageNameIndex)
							} else {
								switch {
								case len(imageInDb) == 0:
									logger.Debug("Запись не найдена в DB. Необходимо создать запись в DB и обновить картинку в WOO")
									imageRowDb.Status = sql.NullString{
										String: image.IMAGE_STATUS_NEED_UPDATE_BY_NOT_FOUND_IN_DB,
										Valid:  true,
									}
									err := imageRowDb.UpdateByIdentRKAndPos(db)
									if err != nil {
										return err
									}
								case len(imageInDb) == 1:
									logger.Debug("Блюдо найдено в DB. Сверяем имя/дату изменения")
									logger.Debugf("Наименование картинки: DB=%s, RK7=%s, File=%s", imageInDb[0].Name.String, imageName, fileInfo.Name())
									logger.Debugf("Дата изменения картинки: DB=%s, File=%s", imageInDb[0].ModTime.String, fileInfo.ModTime().Format(time.RFC3339))
									switch {
									case imageInDb[0].Name.String != imageName:
										logger.Debug("Наименование картинки изменилось. Обновляем картинку в WOO")
										imageRowDb.Status = sql.NullString{
											String: image.IMAGE_STATUS_NEED_UPDATE_BY_DIFF_NAME,
											Valid:  true,
										}
										err := imageRowDb.UpdateByIdentRKAndPos(db)
										if err != nil {
											return err
										}
									case imageInDb[0].ModTime.String != fileInfo.ModTime().Format(time.RFC3339):
										logger.Debug("Дата изменения картинки изменилась. Необходимо обновить картинку")
										imageRowDb.Status = sql.NullString{
											String: image.IMAGE_STATUS_NEED_UPDATE_BY_DIFF_DATE,
											Valid:  true,
										}
										err := imageRowDb.UpdateByIdentRKAndPos(db)
										if err != nil {
											return err
										}
									default:
										logger.Debug("Дата изменения картинки не изменилась. Наименование не изменилось. Необходимо проверить наличие в WOO")
										imageRowDb.Status = sql.NullString{
											String: image.IMAGE_STATUS_NO_NEED_UPDATE,
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
										String: image.IMAGE_STATUS_NEED_UPDATE_BY_FIND_DOUBLE_IN_DB,
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
								String: image.IMAGE_STATUS_FILE_NOT_FOUND,
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
						String: image.IMAGE_STATUS_WOO_NOT_FOUND,
						Valid:  true,
					}
					err = imageRowDb.UpdateByIdentRKAndPos(db)
					if err != nil {
						return errors.Wrap(err, "failed in UpdateByIdentRKAndPos()")
					}
				} else {
					logger.Debug("Блюдо без WOO_ID. Игнорировать")
					imageRowDb.Status = sql.NullString{
						String: image.IMAGE_STATUS_RK7_WOO_ID_NOT_FOUND,
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

// HandlerDbImage 3 этап - Синхронизация DB.Image/DB.ImageFiles с Woo.Product.Image
func HandlerDbImage(db *sqlx.DB) error {

	logger := logging.GetLogger()
	logger.Debug("Start HandlerDbImage")
	defer logger.Debug("End HandlerDbImage")

	handlers := []handler{
		{image.IMAGE_STATUS_WOO_NOT_FOUND, HandlerWooNotFound, "<strong>Ошибки при синхронизации картинок - блюдо RK7 не найдено в WOO, требуется синхронизация:</strong>"}, //+ todo
		{image.IMAGE_STATUS_RK7_IMAGE_NAME_NOT_FOUND, HandlerRk7ImageNameNotFound, "<strong>Ошибки при синхронизации картинок - не указано имя картинка</strong>"},          // todo
		{image.IMAGE_STATUS_FILE_NOT_FOUND, HandlerFileNotFound, "<strong>Ошибки при синхронизации картинок - файл картинки не найден</strong>"},                            // todo
		{image.IMAGE_STATUS_NEED_UPDATE_BY_DIFF_NAME, HandlerNeedUpdateByDiffName, "<strong>Ошибка при обновлении картинки</strong>"},                                       // todo
		{image.IMAGE_STATUS_NEED_UPDATE_BY_DIFF_DATE, HandlerNeedUpdateByDiffDate, "<strong>Ошибка при обновлении картинки</strong>"},                                       // todo
		{image.IMAGE_STATUS_NEED_UPDATE_BY_NOT_FOUND_IN_DB, HandlerNeedUpdateByNotFoundInDb, "<strong>Ошибка при обновлении картинки</strong>"},                             // todo
		{image.IMAGE_STATUS_NEED_UPDATE_BY_FIND_DOUBLE_IN_DB, HandlerNeedUpdateByFindDoubleInDb, "<strong>Ошибка при обновлении картинки</strong>"},                         // todo
		{image.IMAGE_STATUS_NO_NEED_UPDATE, HandlerNoNeedUpdate, "<strong>Ошибка при обновлении картинки</strong>"},                                                         // todo
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

// HandlerVerifyImagesBetweenRk7AndWooAndUpdate сверить картинки между rkeeper/woo и обновить, если есть различии
func HandlerVerifyImagesBetweenRk7AndWooAndUpdate(db *sqlx.DB, status string, message string) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerVerifyImagesBetweenRk7AndWooAndUpdate")
	defer logger.Debug("End HandlerVerifyImagesBetweenRk7AndWooAndUpdate")

	var err error
	var images []*image.Image
	var query string

	logger.Debug("Выполняем поиск записей в таблице Images")
	query = "SELECT * FROM Image WHERE Status=$1 ORDER BY IdentRK, Pos;"
	err = db.Select(&images, query, status)
	logger.Debugf("SELECT:\n%s(%s)", query, status)
	if err != nil {
		return errors.Wrapf(err, "failed SELECT to dbsqlite; query:\n%s(%s)", query, status)
	} else {
		logger.Debugf("Количество полученных строк: %d", len(images))
		if len(images) > 0 {
			m, err := VerifyImagesBetweenRk7AndWooAndUpdate(db, images)
			if err != nil {
				return errors.Wrap(err, "failed in VerifyImagesBetweenRk7AndWooAndUpdate")
			}
			if len(m) > 0 {
				m = append(m, message)
				telegram.SendMessageToTelegramWithLogError(strings.Join(m, "\n"))
			}
		}
	}
	return nil
}

// HandlerWooNotFound IMAGE_STATUS_WOO_NOT_FOUND
// отправляем сообщение об ошибке
func HandlerWooNotFound(db *sqlx.DB, status string, message string) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerWooNotFound")
	defer logger.Debug("End HandlerWooNotFound")

	err := HandlerSendMessage(db, status, message)
	if err != nil {
		return errors.Wrapf(err, "failed in HandlerSendMessage: status=%s, message=%s", status, message)
	} else {
		return nil
	}
}

// HandlerRk7ImageNameNotFound IMAGE_STATUS_RK7_IMAGE_NAME_NOT_FOUND
// проверяем, что RK.IMAGES == WOO.IMAGES
// -если не сходится, то обновляем RK.IMAGES == WOO.IMAGES
// -если сходится, то пропускаем
func HandlerRk7ImageNameNotFound(db *sqlx.DB, status string, message string) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerRk7ImageNameNotFound")
	defer logger.Debug("End HandlerRk7ImageNameNotFound")

	err := HandlerVerifyImagesBetweenRk7AndWooAndUpdate(db, status, message)
	if err != nil {
		return errors.Wrapf(err, "failed in HandlerVerifyImagesBetweenRk7AndWooAndUpdate: status=%s, message=%s", status, message)
	} else {
		return nil
	}
}

// HandlerFileNotFound IMAGE_STATUS_FILE_NOT_FOUND
// проверяем, что RK.IMAGES == WOO.IMAGES
// -если не сходится, то обновляем RK.IMAGES == WOO.IMAGES
// -если сходится, то пропускаем
// отправляем сообщение об ошибке
func HandlerFileNotFound(db *sqlx.DB, status string, message string) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerFileNotFound")
	defer logger.Debug("End HandlerFileNotFound")

	err := HandlerVerifyImagesBetweenRk7AndWooAndUpdate(db, status, message)
	if err != nil {
		return errors.Wrapf(err, "failed in HandlerVerifyImagesBetweenRk7AndWooAndUpdate: status=%s, message=%s", status, message)
	} else {
		err := HandlerSendMessage(db, status, message)
		if err != nil {
			return errors.Wrapf(err, "failed in HandlerSendMessage: status=%s, message=%s", status, message)
		} else {
			return nil
		}
	}
}

// HandlerNeedUpdateByDiffName IMAGE_STATUS_NEED_UPDATE_BY_DIFF_NAME
// проверяем, что RK.IMAGES == WOO.IMAGES
// -если не сходится, то обновляем RK.IMAGES == WOO.IMAGES
// -если сходится, то пропускаем
func HandlerNeedUpdateByDiffName(db *sqlx.DB, status string, message string) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerNeedUpdateByDiffName")
	defer logger.Debug("End HandlerNeedUpdateByDiffName")

	err := HandlerVerifyImagesBetweenRk7AndWooAndUpdate(db, status, message)
	if err != nil {
		return errors.Wrapf(err, "failed in HandlerVerifyImagesBetweenRk7AndWooAndUpdate: status=%s, message=%s", status, message)
	} else {
		return nil
	}
}

// HandlerNeedUpdateByDiffDate IMAGE_STATUS_NEED_UPDATE_BY_DIFF_DATE
// проверяем, что RK.IMAGES == WOO.IMAGES
// -если не сходится, то обновляем RK.IMAGES == WOO.IMAGES
// -если сходится, то пропускаем
func HandlerNeedUpdateByDiffDate(db *sqlx.DB, status string, message string) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerNeedUpdateByDiffDate")
	defer logger.Debug("End HandlerNeedUpdateByDiffDate")

	err := HandlerVerifyImagesBetweenRk7AndWooAndUpdate(db, status, message)
	if err != nil {
		return errors.Wrapf(err, "failed in HandlerVerifyImagesBetweenRk7AndWooAndUpdate: status=%s, message=%s", status, message)
	} else {
		return nil
	}
}

// HandlerNeedUpdateByNotFoundInDb IMAGE_STATUS_NEED_UPDATE_BY_NOT_FOUND_IN_DB
// отправляем сообщение об ошибке
func HandlerNeedUpdateByNotFoundInDb(db *sqlx.DB, status string, message string) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerNeedUpdateByNotFoundInDb")
	defer logger.Debug("End HandlerNeedUpdateByNotFoundInDb")

	err := HandlerSendMessage(db, status, message)
	if err != nil {
		return errors.Wrapf(err, "failed in HandlerSendMessage: status=%s, message=%s", status, message)
	} else {
		return nil
	}
}

// HandlerNeedUpdateByFindDoubleInDb IMAGE_STATUS_NEED_UPDATE_BY_FIND_DOUBLE_IN_DB
// отправляем сообщение об ошибке
func HandlerNeedUpdateByFindDoubleInDb(db *sqlx.DB, status string, message string) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerNeedUpdateByFindDoubleInDb")
	defer logger.Debug("End HandlerNeedUpdateByFindDoubleInDb")

	err := HandlerSendMessage(db, status, message)
	if err != nil {
		return errors.Wrapf(err, "failed in HandlerSendMessage: status=%s, message=%s", status, message)
	} else {
		return nil
	}
}

// HandlerNoNeedUpdate IMAGE_STATUS_NO_NEED_UPDATE
// проверяем, что RK.IMAGES == WOO.IMAGES
// -если не сходится, то обновляем RK.IMAGES == WOO.IMAGES
// -если сходится, то пропускаем
func HandlerNoNeedUpdate(db *sqlx.DB, status string, message string) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerNoNeedUpdate")
	defer logger.Debug("End HandlerNoNeedUpdate")

	err := HandlerVerifyImagesBetweenRk7AndWooAndUpdate(db, status, message)
	if err != nil {
		return errors.Wrapf(err, "failed in HandlerVerifyImagesBetweenRk7AndWooAndUpdate: status=%s, message=%s", status, message)
	} else {
		return nil
	}
}

// HandlerSendMessage отправить сообщение в телеграм
func HandlerSendMessage(db *sqlx.DB, status string, message string) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerSendMessage")
	defer logger.Debug("End HandlerSendMessage")

	var err error
	var images []*image.Image
	var query string

	logger.Debug("Выполняем поиск записей в таблице Images")
	query = "SELECT * FROM Image WHERE Status=$1 ORDER BY IdentRK, Pos;"
	err = db.Select(&images, query, status)
	logger.Debugf("SELECT:\n%s(%s)", query, status)
	if err != nil {
		return errors.Wrapf(err, "failed SELECT to dbsqlite; query:\n%s(%s)", query, status)
	} else {
		logger.Debugf("Количество полученных строк: %d", len(images))
		if len(images) > 0 {
			menu, err := cache.GetMenu()
			if err != nil {
				return errors.Wrap(err, "failed in cache.GetMenu()")
			}
			menuitemsByIdent, err := menu.GetMenuitemsRK7ByIdent()
			if err != nil {
				return errors.Wrap(err, "failed in menu.GetMenuitemsRK7ByIdent()")
			}

			var m []string
			m = append(m, message)
			for _, image := range images {
				var message string
				if menuitem, ok := menuitemsByIdent[image.IdentRK]; ok {
					message = fmt.Sprintf("Блюдо: %s", GetMenuitemDescription(menuitem))
				} else {
					message = fmt.Sprintf("Блюдо с ID=%d не найдено в RK7", image.ID)
				}
				m = append(m, message)
			}
			logger.Debug(strings.Join(m, "\n"))
			telegram.SendMessageToTelegramWithLogError(strings.Join(m, "\n"))
		}
	}
	return nil
}

// CheckImageInWoo Проверяем наличие картинки в WOO.Media
// Используется в 1 этапе синхронизации картинок
func CheckImageInWoo(id int) (int, error) {
	logger := logging.GetLogger()
	logger.Debug("Start CheckImageInWoo")
	defer logger.Debug("End CheckImageInWoo")

	cfg := config.GetConfig()
	url := fmt.Sprintf("%s/wp-json/wp/v2/media/%d", cfg.WOOCOMMERCE.URL, id)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, errors.Wrapf(err, "http.NewRequest")
	}
	req.SetBasicAuth(cfg.WOOCOMMERCE.User, cfg.WOOCOMMERCE.Password)
	resp, err := client.Do(req)
	if err != nil {
		return 0, errors.Wrapf(err, "failed in client.Do")
	}
	switch resp.Status {
	case "200 OK":
		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return 0, errors.Wrapf(err, "ioutil.ReadAll")
		}
		i := new(imageJson)
		err = json.Unmarshal(content, i)
		if err != nil {
			return 0, errors.Wrap(err, "Не удалось выполнить Unmarshal")
		} else {
			return i.Id, nil
		}
	case "404 Not Found":
		return 0, nil
	default:
		return 0, errors.New(fmt.Sprintf("Картинка не закачана. Status: %s", resp.Status))
	}
}

// DeleteImageInWoo Удалить картинку из WOO.Media
// Используется в 1 этапе синхронизации картинок
func DeleteImageInWoo(id int) error {
	logger := logging.GetLogger()
	logger.Debug("Start DeleteImageInWoo")
	defer logger.Debug("End DeleteImageInWoo")

	cfg := config.GetConfig()
	url := fmt.Sprintf("%s/wp-json/wp/v2/media/%d", cfg.WOOCOMMERCE.URL, id)

	client := &http.Client{}
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return errors.Wrapf(err, "http.NewRequest")
	}

	q := req.URL.Query()
	q.Add("force", "true")
	req.URL.RawQuery = q.Encode()
	req.SetBasicAuth(cfg.WOOCOMMERCE.User, cfg.WOOCOMMERCE.Password)
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrapf(err, "failed in client.Do")
	}
	switch resp.Status {
	case "200 OK":
		logger.Debug("Картинка удалена успешно")
		return nil
	default:
		return errors.New(fmt.Sprintf("Картинка не закачана. Status: %s", resp.Status))
	}
}

// UploadImageAndUpdateDb Закачать файл картинки из папки в Woo.Media и обновить в DB поля WOO_ID/ModTime
// Используется в 1 этапе синхронизации картинок
func UploadImageAndUpdateDb(db *sqlx.DB, file fs.FileInfo) error {
	logger := logging.GetLogger()
	logger.Debug("Start UploadImageAndUpdateDb")
	defer logger.Debug("End UploadImageAndUpdateDb")
	filename := file.Name()
	modTime := file.ModTime().Format(time.RFC3339)
	cfg := config.GetConfig()
	url := fmt.Sprintf("%s/wp-json/wp/v2/media", cfg.WOOCOMMERCE.URL)
	path := fmt.Sprintf("%s/%s", cfg.IMAGESYNC.Path, filename)

	data, err := os.Open(path)
	if err != nil {
		return errors.Wrapf(err, "failed in os.Open")
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, data)
	if err != nil {
		return errors.Wrapf(err, "http.NewRequest")
	}
	req.Header.Set("Content-Type", "image/jpeg")
	req.Header.Set("Content-Disposition", fmt.Sprintf(`form-data; filename="%s"`, filename))

	req.SetBasicAuth(cfg.WOOCOMMERCE.User, cfg.WOOCOMMERCE.Password)
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrapf(err, "failed in client.Do")
	}
	switch resp.Status {
	case "201 Created":
		logger.Debug("Картинка закачана успешно")
		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrapf(err, "ioutil.ReadAll")
		}
		i := new(imageJson)
		err = json.Unmarshal(content, i)
		if err != nil {
			return errors.Wrap(err, "Не удалось выполнить Unmarshal")
		} else {
			query := "UPDATE ImageFile SET IdentWOO=:IdentWOO, ModTime=:ModTime WHERE Name=:Name;"
			logger.Debugf("UPDATE:\n%s(%d, %s, %s)", query, i.Id, filename, modTime)
			_, err := db.NamedExec(query,
				map[string]interface{}{
					"Name":     filename,
					"IdentWOO": i.Id,
					"ModTime":  modTime,
				})
			if err != nil {
				return errors.Wrap(err, "failed in db.NamedExec")
			} else {
				return nil
			}
		}
	default:
		return errors.New(fmt.Sprintf("Картинка не закачана. Status: %s", resp.Status))
	}
}

// AddImageToDb Закачать файл картинки из папки в Woo.Media и обновить в DB поля WOO_ID/ModTime
// Используется в 1 этапе синхронизации картинок
func AddImageToDb(db *sqlx.DB, file fs.FileInfo) error {
	logger := logging.GetLogger()
	logger.Debug("Start AddImageToDb")
	defer logger.Debug("End AddImageToDb")

	var err error

	tx := db.MustBegin()
	defer func() {
		if err != nil {
			logger.Error(err)
			err := tx.Rollback()
			if err != nil {
				logger.Errorf("failed in Rollback(); %v", err)
				return
			} else {
				logger.Info("Rollback() is done")
			}
		}
	}()
	query := "INSERT INTO ImageFile (Name, ModTime) VALUES ($1, $2);"
	logger.Debugf("INSERT:\n%s(%s, %s)", query, file.Name(), file.ModTime().Format(time.RFC3339))
	tx.MustExec(query, file.Name(), file.ModTime().Format(time.RFC3339))
	err = tx.Commit()
	if err != nil {
		return errors.Wrapf(err, "failed INSERT to dbsqlite; query:\n%s(%s, %s)", query, file.Name(), file.ModTime().Format(time.RFC3339))
	} else {
		logger.Info("Commit() is done")
		return nil
	}
}

// VerifyImagesBetweenRk7AndWooAndUpdate Сверить картинки и если не сходится, то обновить
// Используется в 3 этапе синхронизации картинок
func VerifyImagesBetweenRk7AndWooAndUpdate(db *sqlx.DB, images []*image.Image) (m []string, err error) {
	logger := logging.GetLogger()
	logger.Debug("Start VerifyImagesBetweenRk7AndWooAndUpdate")
	defer logger.Debug("End VerifyImagesBetweenRk7AndWooAndUpdate")

	menu, err := cache.GetMenu()
	if err != nil {
		return nil, errors.Wrap(err, "failed in cache.GetMenu()")
	}
	menuitemsRK7ByIdent, err := menu.GetMenuitemsRK7ByIdent()
	if err != nil {
		return nil, errors.Wrap(err, "failed in menu.GetMenuitemsRK7ByIdent()")
	}
	productsWooByID, err := menu.GetProductsWooByID()
	if err != nil {
		return nil, errors.Wrap(err, "failed in menu.GetProductsWooByID()")
	}

	logger.Debug("Формируем map по IdentRK, чтобы получить уникальные IdentRK с данной ошибкой")
	imagesMap := make(map[int][]image.Image)
	for _, imageForItem := range images {
		if _, ok := imagesMap[imageForItem.IdentRK]; ok {
			imagesMap[imageForItem.IdentRK] = append(imagesMap[imageForItem.IdentRK], *imageForItem)
		} else {
			imagesMap[imageForItem.IdentRK] = []image.Image{*imageForItem}
		}
	}

	if len(imagesMap) > 0 {
		logger.Debugf("Всего блюд с указанной ошибкой = %d", len(imagesMap))
		for imageIdentRK, image := range imagesMap {
			logger.Debugf("IdentRK: %d; Image: %v", imageIdentRK, image)

			apiWoo := wooapi.GetAPI()
			//сверка

			if menuitem, ok := menuitemsRK7ByIdent[imageIdentRK]; ok {
				if menuitem.WOO_ID != 0 {
					if product, ok := productsWooByID[menuitem.WOO_ID]; ok {
						// формируем list db.images
						// сверяем db.images==woo.images
						// если расхождение то обновляем woo.images

						query := `SELECT Image.IdentRK, Image.Name, Image.Pos, Image.Status, I.ModTime, I.IdentWOO FROM Image
         LEFT JOIN ImageFile I on Image.Name = I.Name
         WHERE Status in ('IMAGE_STATUS_NEED_UPDATE_BY_DIFF_NAME',
                          'IMAGE_STATUS_NEED_UPDATE_BY_DIFF_DATE',
                          'IMAGE_STATUS_NEED_UPDATE_BY_NOT_FOUND_IN_DB',
                          'IMAGE_STATUS_NEED_UPDATE_BY_FIND_DOUBLE_IN_DB',
                          'IMAGE_STATUS_NO_NEED_UPDATE'
                         ) and Image.Name IS NOT NULL and Image.IdentRK == $1
         ORDER BY IdentRK, Pos;`

						var imagesSelectJoinInDb []ImageSelectJoin
						err = db.Select(&imagesSelectJoinInDb, query, imageIdentRK)
						if err != nil {
							return nil, errors.Wrapf(err, "failed SELECT query %s(%d)", query, imageIdentRK)
						}

						if len(imagesSelectJoinInDb) == len(product.Images) {
							//сравнить
							for i, _ := range imagesSelectJoinInDb {
								if int(imagesSelectJoinInDb[i].IdentWOO.Int32) != product.Images[i].Id {
									//выполнить update
									logger.Debugf("Добавляем картинку для продукта: ID:%d, Name:%s", product.ID, product.Name)
									recoveryImages := productsWooByID[menuitem.WOO_ID].Images
									productsWooByID[menuitem.WOO_ID].Images = make([]models2.ProductImage, 0)
									for _, imageAdd := range imagesSelectJoinInDb {
										logger.Debugf("Картинка: ID:%d, Name:%s, IdentRK: %d", imageAdd.IdentWOO.Int32, imageAdd.Name.String, imageAdd.IdentRK)
										productsWooByID[menuitem.WOO_ID].Images = append(productsWooByID[menuitem.WOO_ID].Images,
											models2.ProductImage{Id: int(imageAdd.IdentWOO.Int32)})
									}
									_, err := apiWoo.ProductUpdate(productsWooByID[menuitem.WOO_ID])
									if err != nil {
										m = append(m, fmt.Sprintf("Картинка не обновилась у продукт(rkid=%d, wooid=%d, wooname=%d) не найдено в WOO",
											imageIdentRK, productsWooByID[menuitem.WOO_ID].ID, productsWooByID[menuitem.WOO_ID].Name))
										productsWooByID[menuitem.WOO_ID].Images = recoveryImages
									} else {
										logger.Debug("Картинка обновлена")
									}
									break
								}
							}
						} else {
							//выполнить update
							logger.Debugf("Добавляем картинку для продукта: ID:%d, Name:%s", product.ID, product.Name)
							recoveryImages := productsWooByID[menuitem.WOO_ID].Images
							productsWooByID[menuitem.WOO_ID].Images = make([]models2.ProductImage, 0)
							for _, imageAdd := range imagesSelectJoinInDb {
								logger.Debugf("Картинка: ID:%d, Name:%s, IdentRK: %d", imageAdd.IdentWOO.Int32, imageAdd.Name.String, imageAdd.IdentRK)
								productsWooByID[menuitem.WOO_ID].Images = append(productsWooByID[menuitem.WOO_ID].Images,
									models2.ProductImage{Id: int(imageAdd.IdentWOO.Int32)})
							}
							_, err := apiWoo.ProductUpdate(productsWooByID[menuitem.WOO_ID])
							if err != nil {
								m = append(m, fmt.Sprintf("Картинка не обновилась у продукт(rkid=%d, wooid=%d, wooname=%d) не найдено в WOO",
									imageIdentRK, productsWooByID[menuitem.WOO_ID].ID, productsWooByID[menuitem.WOO_ID].Name))
								productsWooByID[menuitem.WOO_ID].Images = recoveryImages
							} else {
								logger.Debug("Картинка обновлена")
							}
						}
					} else {
						m = append(m, fmt.Sprintf("Блюдо(id=%d, wooid=%d) не найдено в WOO", imageIdentRK, menuitem.WOO_ID))
					}
				} else {
					m = append(m, fmt.Sprintf("Блюдо(id=%d) без WOO_ID", imageIdentRK))
				}
			} else {
				m = append(m, fmt.Sprintf("Блюдо(id=%d) не найдено в RK7", imageIdentRK))
			}
		}
	}
	return nil, nil
}

// GetMenuitemDescription Сформировать строку с блюдом
// Используется в 3 этапе синхронизации картинок
func GetMenuitemDescription(m *models.MenuitemItem) string {
	return fmt.Sprintf("ID=%d, Name=%s, LongName=%s, Цена=%d, WooID=%d", m.ItemIdent, m.Name, m.WOO_LONGNAME, m.PRICETYPES, m.WOO_ID)
}
