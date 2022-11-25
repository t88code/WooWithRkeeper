package sync

import (
	"WooWithRkeeper/internal/cache"
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/internal/database"
	modelsRK7API "WooWithRkeeper/internal/rk7api/models"
	"WooWithRkeeper/internal/telegram"
	"WooWithRkeeper/internal/wooapi"
	"WooWithRkeeper/internal/wooapi/models"
	"WooWithRkeeper/pkg/logging"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type productImage struct {
	Id              int    `json:"id"`
	DateCreated     string `json:"date_created"`
	DateCreatedGmt  string `json:"date_created_gmt"`
	DateModified    string `json:"date_modified"`
	DateModifiedGmt string `json:"date_modified_gmt"`
	Src             string `json:"src"`
	Name            string `json:"name"`
	Alt             string `json:"alt"`
}

type menuitemError struct {
	*modelsRK7API.MenuitemItem
	errorText string
}

func (me menuitemError) Error() string {
	dish := fmt.Sprintf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d, RK_Image: %s",
		me.Name, me.WOO_LONGNAME, me.ItemIdent, me.WOO_ID, me.WOO_PARENT_ID, me.Status, me.PRICETYPES, me.WOO_IMAGE_NAME)

	return fmt.Sprintf("%s; error: %s", dish, me.errorText)
}

func SyncImages() error {

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

	var resultSyncAll []string
	var resultSyncError []string

	logger.Debug("Получаем меню из RK7 и WOO")
	menu, err := cache.GetMenu()
	if err != nil {
		return err
	}

	menuitems, err := menu.GetMenuitems()
	if err != nil {
		return err
	}

	if len(menuitems) == 0 {
		err = menu.RefreshMenuitems()
		if err != nil {
			return err
		}
		menuitems, err = menu.GetMenuitems()
		if err != nil {
			return err
		}
	}

	productsWooByID, err := menu.GetProductsWooByID()
	if err != nil {
		return err
	}

	if len(productsWooByID) == 0 {
		err = menu.RefreshProducts()
		if err != nil {
			return err
		}

		productsWooByID, err = menu.GetProductsWooByID()
		if err != nil {
			return err
		}
	}

	logger.Debug("Запущен процесс обновления картинок блюд")

	// блюда RK7
	var menuitemsActive int                                  // активные
	var menuitemsNotActive int                               // не активные
	var menuitemsPriceNotDefine []*modelsRK7API.MenuitemItem // не указана цена - do Отправить сообщение

	var menuitemsFoundInWoo []*modelsRK7API.MenuitemItem    // найдено в WOO
	var menuitemsNotFoundInWoo []*modelsRK7API.MenuitemItem // не найдено в WOO - do Отправить сообщение

	var menuitemsWooIsNull []*modelsRK7API.MenuitemItem     // WooID не указан - do Отправить сообщение
	var menuitemsImageNameNull []*modelsRK7API.MenuitemItem // WooImageName не указан  - do Отправить сообщение и обнулить картинку в WOO
	var menuitemsImageNotFound []*modelsRK7API.MenuitemItem // WooImageName указан, но картинка не найдена - do Отправить сообщение и обнулить картинку в WOO
	var menuitemsNonJpgFile []*modelsRK7API.MenuitemItem    // картинка формата не jpg - todo сообщить

	var menuitemsNeedUpdateInWooByDate []*modelsRK7API.MenuitemItem   // дата картинки изменилась - do Обновить/создать картинку в WOO
	var menuitemsNoNeedUpdateInWooByDate []*modelsRK7API.MenuitemItem // дата картинки не изменилась - do Проверить картинку в WOO

	var menuitemsNeedAddInDB []*modelsRK7API.MenuitemItem // нет записи в DB - do Обновить/создать картинку в WOO и добавить запись в DB
	var menuitemsDubleInDB []*modelsRK7API.MenuitemItem   // дубли в DB - do Отправить сообщение

LoopOneStage:
	for i, menuitem := range menuitems {
		logger.Debug("--------------------------------------")
		dish := fmt.Sprintf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d, RK_Image: %s",
			menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES, menuitem.WOO_IMAGE_NAME)
		logger.Debugf(dish)

		for _, ignoreIdent := range cfg.RK7.MenuitemIdentIgnore {
			if menuitem.ItemIdent == ignoreIdent {
				logger.Debug("Игнорируем по настройкам конфига")
				continue LoopOneStage
			}
		}

		if menuitem.Status != 3 {
			logger.Debug("Не активное блюдо. Пропускаем. Синхронизация меню отключила блюдо в WOO")
			menuitemsNotActive++
		} else {
			logger.Debug("Блюдо активное. Проверяем цену")
			menuitemsActive++
			if menuitem.PRICETYPES == 9223372036854775807 {
				logger.Debug("Блюдо без указанной цены. Сообщаем и пропускаем.")
				menuitemsPriceNotDefine = append(menuitemsPriceNotDefine, menuitems[i])
			} else {
				if _, found := productsWooByID[menuitem.WOO_ID]; found {
					logger.Debug("Блюдо найдено в WOO. Проверяем наименование картинки")
					menuitemsFoundInWoo = append(menuitemsFoundInWoo, menuitems[i])
					if menuitem.WOO_IMAGE_NAME == "" {
						logger.Debug("Не указано наименование картинки. Отправить сообщение и обнулить картинку в WOO")
						menuitemsImageNameNull = append(menuitemsImageNameNull, menuitems[i])
					} else {
						logger.Debugf("Указано наименование картинки %s", menuitem.WOO_IMAGE_NAME)
						logger.Debugf("Проверяем наличие картинки в папке %s", cfg.IMAGESYNC.Path)
						path := fmt.Sprintf("%s%s", cfg.IMAGESYNC.Path, menuitem.WOO_IMAGE_NAME)
						if fileInfo, err := os.Stat(path); !os.IsNotExist(err) {
							logger.Debug("Файл картинки существует. Проверяем, что это jpg")
							matchJPG, _ := regexp.MatchString(".jpg$", fileInfo.Name())

							if matchJPG {
								logger.Debug("Сверяем дату изменения c сохраненной в DB")
								var menuitemDBs []database.Menuitem
								query := fmt.Sprintf(`SELECT ID, IdentRK, ImageModTime FROM Menuitem WHERE IdentRK=%d`, menuitem.ItemIdent)
								err = db.Select(&menuitemDBs, query)
								if err != nil {
									return errors.Wrapf(err, "failed SELECT to dbsqlite; dish %s", dish)
								} else {
									if len(menuitemDBs) > 0 {
										logger.Debug("Блюдо найдено в DB. Сверяем дату изменения.")
										modTimeFile := fileInfo.ModTime().Format(time.RFC3339)
										modTimeDB := menuitemDBs[0].IMAGE_MOD_TIME
										logger.Debugf("Дата изменения картинки в папке: %s", modTimeFile)
										logger.Debugf("Дата изменения картинки в DB: %s", modTimeDB)
										if modTimeFile != modTimeDB {
											logger.Debug("Дата изменения картинки изменилась. Необходимо обновить картинку")
											menuitemsNeedUpdateInWooByDate = append(menuitemsNeedUpdateInWooByDate, menuitems[i])
										} else {
											logger.Debug("Дата изменения картинки не изменилась. Сверяем наименование картинки между WOO и RK7")
											menuitemsNoNeedUpdateInWooByDate = append(menuitemsNoNeedUpdateInWooByDate, menuitems[i])
										}
									} else if len(menuitemDBs) > 1 {
										errText := fmt.Sprintf("Недопустимое количество блюд в DB = %d > 1", len(menuitemDBs))
										logger.Debugf(errText)
										menuitemsDubleInDB = append(menuitemsDubleInDB, menuitems[i])
									} else {
										logger.Debug("Запись не найдена в DB. Необходимо создать запись в DB и обновить картинку в WOO")
										menuitemsNeedAddInDB = append(menuitemsNeedAddInDB, menuitems[i])
									}
								}
							} else {
								errText := fmt.Sprintf("Картинка %s формата не jpg", fileInfo.Name())
								logger.Debugf(errText)
								menuitemsNonJpgFile = append(menuitemsNonJpgFile, menuitems[i])
							}

						} else {
							logger.Debug("Файл картинки не найден. Отправить сообщение и обнулить картинку в WOO")
							menuitemsImageNotFound = append(menuitemsImageNotFound, menuitems[i])
						}
					}
				} else {
					if menuitem.WOO_ID != 0 {
						logger.Debug("Блюдо не найдено в WOO. Отправить сообщение")
						menuitemsNotFoundInWoo = append(menuitemsNotFoundInWoo, menuitems[i])
					} else {
						logger.Debug("Блюдо без WOO_ID. Игнорировать")
						menuitemsWooIsNull = append(menuitemsWooIsNull, menuitems[i])
					}
				}
			}
		}
	}

	logger.Debug("Блюда WOO:")
	logger.Debugf("Всего: %d", len(productsWooByID))

	logger.Debug("Блюда RK7:")
	logger.Debugf("Всего: %d", len(menuitems))
	logger.Debugf("Активные: %d", menuitemsActive)                                //++
	logger.Debugf("Не активные: %d", menuitemsNotActive)                          //++
	logger.Debugf("Не указана цена - сообщаем: %d", len(menuitemsPriceNotDefine)) //++
	logger.Debugf("Игнорировано: %d", len(cfg.RK7.MenuitemIdentIgnore))

	logger.Debugf("Блюда найдено в WOO: %d", len(menuitemsFoundInWoo))                  //++
	logger.Debugf("Блюдо не найдено в WOO - сообщить: %d", len(menuitemsNotFoundInWoo)) //++
	logger.Debugf("Блюдо без WOO_ID - сообщить: %d", len(menuitemsWooIsNull))           //++

	logger.Debugf("Необходимо сообщить и обнулить картинку в WOO, причина - пустое поле WOO_IMAGE_NAME: %d", len(menuitemsImageNameNull))                //++
	logger.Debugf("Необходимо сообщить и обнулить картинку в WOO, причина - есть поле WOO_IMAGE_NAME, но нет картинки: %d", len(menuitemsImageNotFound)) //++
	logger.Debugf("Необходимо сообщить в WOO, причина - картинка формата не jpg: %d", len(menuitemsNonJpgFile))

	logger.Debugf("Необходимо обновить картинку в WOO, причина - дата изменилась: %d", len(menuitemsNeedUpdateInWooByDate))       //++++
	logger.Debugf("Необходимо проверить картинку в WOO, причина - дата не изменилась: %d", len(menuitemsNoNeedUpdateInWooByDate)) //++
	logger.Debugf("Необходимо обновить картинку в WOO, причина - нет записи в DB: %d", len(menuitemsNeedAddInDB))                 //++
	logger.Debugf("Дубли в DB - сообщить: %d", len(menuitemsDubleInDB))                                                           //++

	//++++++
	if len(menuitemsPriceNotDefine) > 0 {
		logger.Debugf("Найдены блюда без цены, отправляем сообщение")
		resultSyncAll = append(resultSyncAll, "<strong>Блюда RK7 без цены:</strong>")
		for _, menuitem := range menuitemsPriceNotDefine {
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

	//++++
	if len(menuitemsWooIsNull) > 0 {
		logger.Debugf("Блюда RK7, WOO_ID не указан, отправляем сообщение")
		resultSyncAll = append(resultSyncAll, "<strong>Блюда RK7, WOO_ID не указан:</strong>")
		for _, menuitem := range menuitemsWooIsNull {
			dish := fmt.Sprintf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d, RK_Image: %s",
				menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES, menuitem.WOO_IMAGE_NAME)
			resultSyncAll = append(resultSyncAll, dish)
		}
	}

	//++++
	if len(menuitemsImageNameNull) > 0 {
		logger.Debugf("Блюда RK7, WOO_IMAGE_NAME не указан, обнуляем в WOO и отправляем сообщение")
		resultSyncAll = append(resultSyncAll, "<strong>Блюда RK7, WOO_IMAGE_NAME не указан:</strong>")
		for i, menuitem := range menuitemsImageNameNull {
			dish := fmt.Sprintf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d, RK_Image: %s",
				menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES, menuitem.WOO_IMAGE_NAME)
			var messageText string
			err := NulledImageInWoo(menuitemsImageNameNull[i])
			if err != nil {
				messageText = fmt.Sprintf("%s; Не удалось обнулить в WOO: %v", dish, err)
			} else {
				messageText = fmt.Sprintf("%s; Успешно обнулено в WOO", dish)
			}
			resultSyncAll = append(resultSyncAll, messageText)
		}
	}

	//++++
	if len(menuitemsImageNotFound) > 0 {
		logger.Debugf("Блюда RK7, WOO_IMAGE_NAME указан, картинка не найдена в папке, обнуляем в WOO и отправляем сообщение")
		resultSyncAll = append(resultSyncAll, "<strong>Блюда RK7, картинки не найдены в папке:</strong>")
		resultSyncError = append(resultSyncError, "<strong>Блюда RK7, картинки не найдены в папке:</strong>")

		for i, menuitem := range menuitemsImageNotFound {
			dish := fmt.Sprintf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d, RK_Image: %s",
				menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES, menuitem.WOO_IMAGE_NAME)
			var messageText string
			err := NulledImageInWoo(menuitemsImageNotFound[i])
			if err != nil {
				messageText = fmt.Sprintf("%s; Не удалось обнулить в WOO: %v", dish, err)
			} else {
				messageText = fmt.Sprintf("%s; Успешно обнулено в WOO", dish)
			}
			resultSyncAll = append(resultSyncAll, messageText)
			resultSyncError = append(resultSyncError, messageText)
		}
	}

	//++++
	if len(menuitemsNeedUpdateInWooByDate) > 0 {
		logger.Debugf("Блюда RK7, дата картинки изменилась, обновляем/создаем картинку в WOO, кол-во %d", len(menuitemsNeedUpdateInWooByDate))
		resultSyncAll = append(resultSyncAll, fmt.Sprintf("<strong>Блюда RK7, картинки изменились, обновляем в WOO, кол-во %d:</strong>", len(menuitemsNeedUpdateInWooByDate)))
		var messageText string
		var failedUpdateCount int
		var mErrors []string
		mErrors = append(mErrors, fmt.Sprintf("<strong>Блюда RK7, картинки изменились, обновляем в WOO, кол-во %d:</strong>", len(menuitemsNeedUpdateInWooByDate)))
		for i, menuitem := range menuitemsNeedUpdateInWooByDate {
			dish := fmt.Sprintf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d, RK_Image: %s",
				menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES, menuitem.WOO_IMAGE_NAME)
			err := CheckImageInWoo(menuitemsNeedUpdateInWooByDate[i])
			if err != nil {
				switch err.Error() {
				case ERROR_IMAGE_NOT_FOUND_IN_WOO:
					messageText = fmt.Sprintf("%s; %v", dish, err)
					mErrors = append(mErrors, messageText)
					failedUpdateCount++
				case ERROR_IMAGE_NOT_FOUND_IN_WOO_IMAGE:
					err = AddImageInWoo(menuitemsNeedUpdateInWooByDate[i])
					if err != nil {
						messageText = fmt.Sprintf("%s; Не удалось создать в Woo: %v", dish, err)
						mErrors = append(mErrors, messageText)
						failedUpdateCount++
					} else {
						logger.Debug("Image успешно создан в Woo. Обновляем запись в DB")
						path := fmt.Sprintf("%s%s", cfg.IMAGESYNC.Path, menuitem.WOO_IMAGE_NAME)
						if fileInfo, err := os.Stat(path); !os.IsNotExist(err) {
							err := UpdateImageInDB(menuitemsNeedUpdateInWooByDate[i], fileInfo.ModTime())
							if err != nil {
								messageText = fmt.Sprintf("%s; Не удалось обновить в DB: %v", dish, err)
								mErrors = append(mErrors, messageText)
								failedUpdateCount++
							} else {
								messageText = "Image успешно обновлен в DB"
								logger.Debug(messageText)
								resultSyncAll = append(resultSyncAll, messageText)
								continue
							}
						} else {
							messageText = fmt.Sprintf("%s; Не удалось найти картинку в папке ", dish)
							mErrors = append(mErrors, messageText)
							failedUpdateCount++
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
				logger.Debug("Картинка существует, обновляем")
				err = UpdateImageInWoo(menuitemsNeedUpdateInWooByDate[i])
				if err != nil {
					messageText = fmt.Sprintf("%s; Не удалось обновить в WOO: %v", dish, err)
					logger.Error(messageText)
					mErrors = append(mErrors, messageText)
					failedUpdateCount++
				} else {
					logger.Debug("Image успешно обновлен в WOO. Обновляем запись в DB")
					path := fmt.Sprintf("%s%s", cfg.IMAGESYNC.Path, menuitem.WOO_IMAGE_NAME)
					if fileInfo, err := os.Stat(path); !os.IsNotExist(err) {
						err := UpdateImageInDB(menuitemsNeedUpdateInWooByDate[i], fileInfo.ModTime())
						if err != nil {
							messageText = fmt.Sprintf("%s; Не удалось обновить в DB: %v", dish, err)
							mErrors = append(mErrors, messageText)
							failedUpdateCount++
						} else {
							messageText = "Image успешно обновлен в DB"
							logger.Debug(messageText)
							resultSyncAll = append(resultSyncAll, messageText)
							continue
						}
					} else {
						messageText = fmt.Sprintf("%s; Не удалось найти картинку в папке ", dish)
						mErrors = append(mErrors, messageText)
					}
				}
				resultSyncAll = append(resultSyncAll, messageText)
				logger.Debug(messageText)
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

	//++++
	if len(menuitemsNoNeedUpdateInWooByDate) > 0 {
		logger.Debugf("Блюда RK7, дата картинки не изменилась, проверяем в WOO, что картинка существует, кол-во %d", len(menuitemsNoNeedUpdateInWooByDate))

		var messageText string
		var failedUpdateCount int

		var mErrors []string
		mErrors = append(mErrors, fmt.Sprintf("<strong>Блюда RK7, дата картинок не изменилась, проверяем в WOO, кол-во %d:</strong>", len(menuitemsNoNeedUpdateInWooByDate)))
		for i, menuitem := range menuitemsNoNeedUpdateInWooByDate {
			dish := fmt.Sprintf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d, RK_Image: %s",
				menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES, menuitem.WOO_IMAGE_NAME)
			err := CheckImageInWoo(menuitemsNoNeedUpdateInWooByDate[i])
			if err != nil {
				switch err.Error() {
				case ERROR_IMAGE_NOT_FOUND_IN_WOO:
					messageText = fmt.Sprintf("%s; %v", dish, err)
					mErrors = append(mErrors, messageText)
					failedUpdateCount++
				case ERROR_IMAGE_NOT_FOUND_IN_WOO_IMAGE:
					err = AddImageInWoo(menuitemsNoNeedUpdateInWooByDate[i])
					if err != nil {
						messageText = fmt.Sprintf("%s; Не удалось создать в WOO: %v", dish, err)
						mErrors = append(mErrors, messageText)
						failedUpdateCount++
					} else {
						logger.Debug("Image успешно создан в Woo. Обновляем запись в DB")
						path := fmt.Sprintf("%s%s", cfg.IMAGESYNC.Path, menuitem.WOO_IMAGE_NAME)
						if fileInfo, err := os.Stat(path); !os.IsNotExist(err) {
							err := UpdateImageInDB(menuitemsNoNeedUpdateInWooByDate[i], fileInfo.ModTime())
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
			messageText = fmt.Sprintf("Картинки успешно проверены в WOO, кол-во %d", len(menuitemsNoNeedUpdateInWooByDate))
		} else if failedUpdateCount < len(menuitemsNoNeedUpdateInWooByDate) {
			messageText = fmt.Sprintf("Остальные картинки успешно проверены в WOO, кол-во %d", len(menuitemsNoNeedUpdateInWooByDate)-failedUpdateCount)
		} else if failedUpdateCount == len(menuitemsNoNeedUpdateInWooByDate) {
			messageText = fmt.Sprintf("Ни одна картинка не проверена в WOO, кол-во %d", len(menuitemsNoNeedUpdateInWooByDate)-failedUpdateCount)
		} else {
			messageText = "Неизвестная ошибка при проверке картинки в WOO"
		}
		logger.Debug(messageText)
		resultSyncAll = append(resultSyncAll, messageText)
	}

	//++++
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
			logger.Debug("Обновляем картинку в DB")
			path := fmt.Sprintf("%s%s", cfg.IMAGESYNC.Path, menuitem.WOO_IMAGE_NAME)
			if fileInfo, err := os.Stat(path); !os.IsNotExist(err) {
				err := UpdateImageInDB(menuitemsNeedAddInDB[i], fileInfo.ModTime())
				if err != nil {
					failedUpdateCount++
					messageText = fmt.Sprintf("Не удалось обновить картинку в DB; %s; %v", dish, err)
					mErrors = append(mErrors, messageText)
				} else {
					logger.Debug("Успешно обновлена картинка в DB")
					logger.Debug("Приступаем к добавлению картинки в WOO")
					err := CheckImageInWoo(menuitemsNeedAddInDB[i])
					if err != nil {
						switch err.Error() {
						case ERROR_IMAGE_NOT_FOUND_IN_WOO:
							messageText = fmt.Sprintf("%s; %v", dish, err)
							failedUpdateCount++
							mErrors = append(mErrors, messageText)
						case ERROR_IMAGE_NOT_FOUND_IN_WOO_IMAGE:
							err = AddImageInWoo(menuitemsNeedAddInDB[i])
							if err != nil {
								messageText = fmt.Sprintf("%s; Не удалось создать картинку в WOO: %v", dish, err)
								failedUpdateCount++
								mErrors = append(mErrors, messageText)
							} else {
								logger.Debug("Image успешно создан в WOO")
								messageText = fmt.Sprintf("%s; Картинка успешно добавлена в WOO: %v", dish, err)
							}
						case ERROR_IMAGE_CHECK_UNDEFINE:
							messageText = fmt.Sprintf("%s; %v", dish, err)
							failedUpdateCount++
							mErrors = append(mErrors, messageText)
						case ERROR_IMAGE_CHECK_ERROR_CAST:
							messageText = fmt.Sprintf("%s; %v", dish, err)
							failedUpdateCount++
							mErrors = append(mErrors, messageText)
						default:
							messageText = fmt.Sprintf("%s; %s", dish, "Неизвестная ошибка")
							failedUpdateCount++
							mErrors = append(mErrors, messageText)
						}
						resultSyncAll = append(resultSyncAll, messageText)
						logger.Debug(messageText)
					} else {
						logger.Debug("Картинка существует, обновляем")
						err = UpdateImageInWoo(menuitemsNeedAddInDB[i])
						if err != nil {
							messageText = fmt.Sprintf("%s; Не удалось обновить в WOO: %v", dish, err)
							logger.Error(messageText)
							resultSyncAll = append(resultSyncAll, messageText)
							failedUpdateCount++
							mErrors = append(mErrors, messageText)
						} else {
							logger.Debug("Image успешно обновлен в WOO")
						}
					}
				}
			} else {
				messageText = fmt.Sprintf("%s; Не удалось найти картинку в папке ", dish)
				mErrors = append(mErrors, messageText)
				resultSyncAll = append(resultSyncAll, messageText)
			}
		}

		if failedUpdateCount == 0 {
			messageText = fmt.Sprintf("Картинки успешно обновлены в WOO, кол-во %d", len(menuitemsNeedAddInDB))
		} else if failedUpdateCount < len(menuitemsNeedAddInDB) {
			messageText = fmt.Sprintf("Остальные картинки успешно обновлены в WOO, кол-во %d", len(menuitemsNeedAddInDB)-failedUpdateCount)
		} else {
			messageText = fmt.Sprintf("Ни одна картинка не обновлена, кол-во %d", len(menuitemsNeedAddInDB))
		}
		resultSyncAll = append(resultSyncAll, messageText)

		if len(mErrors) > 1 {
			resultSyncError = append(resultSyncError, mErrors...)
		}

	}

	//++++
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

	//++++
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
	return nil
}

func NulledImageInWoo(menuitem *modelsRK7API.MenuitemItem) error {
	//todo

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
				//status, _ := SendDeleteRequest(fmt.Sprintf("http://new.hotelslovakia.ru/wp-json/wp/v2/media/%d", image.Id)) //todo url
				status, _ := SendDeleteRequest(fmt.Sprintf("%s/wp-json/wp/v2/media/%d", cfg.WOOCOMMERCE.URL, image.Id)) //todo url
				switch status {
				case "201 Created":
					logger.Debug("Картинка создана!")
					os.Exit(44) // todo
				case "200 OK":
					logger.Debug("Картинка удалена!")
					logger.Debug("Обнуляем запись в DB")
					err := NulledImageInDB(menuitem)
					if err != nil {
						logger.Errorf("Ошибка при обнулении записи в DB; %v", err) // todo errors везде
						return err
					} else {
						logger.Debug("Обнулена запись в DB")
						return nil
					}
				default:
					os.Exit(23) // todo
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
	tx.MustExec("DELETE FROM Menuitem WHERE IdentRK = $1;", menuitem.ItemIdent)
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

	if _, found := productsWooByID[menuitem.WOO_ID]; found {

		if len(productsWooByID[menuitem.WOO_ID].Images) == 0 {
			err := AddImageInWoo(menuitem)
			if err != nil {
				logger.Error("UpdateImageInWoo:Картинка не добавлена в продукт в WOO")
				return errors.Wrap(err, "UpdateImageInWoo:Картинка не добавлена в продукт в WOO")
			} else {
				logger.Debug("Картинка успешно добавлена в продукт в WOO")
				return nil
			}
		} else {
			for i, image := range productsWooByID[menuitem.WOO_ID].Images {
				logger.Debug(image)
				//if imageMap, ok := image.(imageS); ok {
				if menuitem.WOO_IMAGE_NAME != "" {
					if image.Alt == menuitem.WOO_IMAGE_NAME {
						logger.Debug("Картинка найдена")

						id := image.Id

						//productsWooByID[menuitem.WOO_ID].Images = append(productsWooByID[menuitem.WOO_ID].Images[:i], productsWooByID[menuitem.WOO_ID].Images[i+1:]...)

						//todo updateImage

						logger.Debug("Обновляем картинку в WOO")

						//err := UpdateImageInWooAPI(menuitem, image.Id)
						//if err != nil {
						//	logger.Error("Картинка не обновлена в продукте в WOO")
						//	return errors.Wrap(err, "Картинка не обновлена в продукте в WOO")
						//} else {
						//	logger.Debug("Картинка успешно обновлена в продукте в WOO")
						//	return nil
						//}
						cfg := config.GetConfig()

						logger.Debugf("Приступаем к удалению файла в WOO ID=%d", id)
						status, _ := SendDeleteRequest(fmt.Sprintf("%s/wp-json/wp/v2/media/%d", cfg.WOOCOMMERCE.URL, id))
						switch status {
						case "201 Created":
							logger.Debug("Картинка создана!")
						case "200 OK":
							productsWooByID[menuitem.WOO_ID].Images = append(productsWooByID[menuitem.WOO_ID].Images[:i], productsWooByID[menuitem.WOO_ID].Images[i+1:]...)
							logger.Debug("Картинка удалена!")
							err := AddImageInWoo(menuitem)
							if err != nil {
								logger.Error("UpdateImageInWoo:AddImageInWoo:Картинка не добавлена в продукт в WOO")
								return errors.Wrap(err, "UpdateImageInWoo:AddImageInWoo:Картинка не добавлена в продукт в WOO")
							} else {
								logger.Debug("Картинка успешно добавлена в продукт в WOO")
								return nil
							}
						default:
							os.Exit(2222) // todo
						}

						//api := wooapi.GetAPI()
						//_, err := api.ProductUpdate(productsWooByID[menuitem.WOO_ID])
						//if err != nil {
						//	logger.Error(err, "Картинка не обновилась у продукта")
						//	return errors.Wrapf(err, "Картинка не обновилась у продукта")
						//} else {
						//	logger.Debug("Картинка удалена у продукта в WOO")
						//	logger.Debug("Приступаем к удалению файла в WOO")
						//	status, _ := SendDeleteRequest(fmt.Sprintf("http://new.hotelslovakia.ru/wp-json/wp/v2/media/%d", id))
						//
						//	fmt.Println(status)
						//	os.Exit(5)
						//	switch status {
						//	case "201 Created":
						//		logger.Debug("Картинка создана!")
						//	case "200 OK":
						//		logger.Debug("Картинка удалена!")
						//	default:
						//		os.Exit(2222)
						//	}
						//
						//	err := AddImageInWoo(menuitem)
						//	if err != nil {
						//		logger.Error("Картинка не добавлена в продукт в WOO")
						//		return errors.Wrap(err, "Картинка не добавлена в продукт в WOO")
						//	} else {
						//		logger.Debug("Картинка успешно добавлена в продукт в WOO")
						//		return nil
						//	}
						//}
					} else {
						logger.Error("Не найдено имя картинки WOO_IMAGE_NAME")
						err := AddImageInWoo(menuitem)
						if err != nil {
							logger.Error("Картинка не добавлена в продукт в WOO")
							return errors.Wrap(err, "Картинка не добавлена в продукт в WOO")
						} else {
							logger.Debug("Картинка успешно добавлена в продукт в WOO")
							return nil
						}
					}
				} else {
					logger.Error("menuitem.WOO_IMAGE_NAME не указан")
					return errors.New("menuitem.WOO_IMAGE_NAME не указан")
				}

				//} else {
				//	logger.Debug("Ошибка приведения. Отправим сообщение об ошибке")
				//	return errors.New(ERROR_IMAGE_CHECK_ERROR_CAST)
				//}
			}
			return errors.New("Неизвестная ошибка при обновлении картинки")
		}
	} else {
		return errors.New("Продукт не найден")
	}
}

func AddImageInWoo(menuitem *modelsRK7API.MenuitemItem) error {
	// todo срань переделать
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
			//imageUrl := imageJ.SourceUrl
			imageID := imageJ.Id
			api := wooapi.GetAPI()

			if _, found := productsWooByID[menuitem.WOO_ID]; found {

				var i models.ProductImage
				//i.Src = imageUrl
				i.Alt = menuitem.WOO_IMAGE_NAME
				i.Id = imageID

				productsWooByID[menuitem.WOO_ID].Images = append(productsWooByID[menuitem.WOO_ID].Images, i)

				_, err := api.ProductUpdate(productsWooByID[menuitem.WOO_ID])
				if err != nil {
					logger.Error(err, "Картинка не обновилась у продукт")
					return errors.Wrapf(err, "Картинка не обновилась у продукт")
				} else {
					logger.Debug("Картинка обновлена")
				}

			} else {
				return errors.New("Продукт не найден")
			}

		}
	case "200":
		logger.Debug("200 okk")
		os.Exit(11111) // todo

	default:
		os.Exit(2222) // todo
	}

	return nil
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

	req.SetBasicAuth("restocrm", "108restocrm108")
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
	client := &http.Client{}

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	q := req.URL.Query()
	q.Add("force", "true")
	req.URL.RawQuery = q.Encode()
	req.SetBasicAuth("restocrm", "108restocrm108")
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

	if len(productsWooByID) == 0 {
		err = menu.RefreshProducts()
		if err != nil {
			return err
		}

		productsWooByID, err = menu.GetProductsWooByID()
		if err != nil {
			return err
		}
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

				//if imageMap, ok := image.(productImage); ok {
				//	if imageMap.Name == menuitem.WOO_IMAGE_NAME || imageMap.Name != "" {
				//		logger.Debug("Картинка найдена")
				//		return nil
				//	}
				//} else {
				//	logger.Debug("Ошибка приведения. Отправим сообщение об ошибке")
				//	return errors.New(ERROR_IMAGE_CHECK_ERROR_CAST)
				//}

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

	var menuitemDBs []database.Menuitem
	query := fmt.Sprintf(`SELECT ID, IdentRK, ImageModTime FROM Menuitem WHERE IdentRK=%d`, menuitem.ItemIdent)
	err = db.Select(&menuitemDBs, query)
	if err != nil {
		return errors.Wrapf(err, "failed SELECT to dbsqlite")
	} else {
		if len(menuitemDBs) == 0 {
			tx := db.MustBegin()
			tx.MustExec("INSERT INTO Menuitem (IdentRK, ImageModTime) VALUES ($1, $2)", menuitem.ItemIdent, modTime.Format(time.RFC3339))
			err := tx.Commit()
			if err != nil {
				return errors.Wrapf(err, "failed INSERT to dbsqlite")
			} else {
				return nil
			}
		} else if len(menuitemDBs) == 1 {
			tx := db.MustBegin()
			tx.MustExec("UPDATE Menuitem SET ImageModTime = $1 WHERE IdentRK = $2", modTime.Format(time.RFC3339), menuitem.ItemIdent)
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
	ERROR_IMAGE_CHECK_UNDEFINE         = "Неопределенная ошибка при поиске картинки в WOO:Image" //todo не используется
	ERROR_IMAGE_CHECK_ERROR_CAST       = "Ошибка приведения"
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

//type imageS struct {
//	Id              int    `json:"id,omitempty"`
//	DateCreated     string `json:"date_created,omitempty"`
//	DateCreatedGmt  string `json:"date_created_gmt,omitempty"`
//	DateModified    string `json:"date_modified,omitempty"`
//	DateModifiedGmt string `json:"date_modified_gmt,omitempty"`
//	Src             string `json:"src,omitempty"`
//	Name            string `json:"name,omitempty"`
//	Alt             string `json:"alt,omitempty"`
//}
