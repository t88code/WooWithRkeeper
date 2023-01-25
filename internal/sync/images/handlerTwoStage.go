package images

import (
	"WooWithRkeeper/internal/cache"
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/pkg/logging"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"time"
)

// HandlerMenuitemsToDbImage 2 этап - Синхронизация Menuitems c DB.Image
func HandlerMenuitemsToDbImage(db *sqlx.DB, imagesSync *[]ImageSync) error {

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

	imageFilesInDbMap, err := GetImageFilesInDbMap(db)
	if err != nil {
		return errors.Wrap(err, "failed in GetImageFilesInDbMap()")
	}

	// блюда RK7
	var menuitemsActive int    // активные - счетчик
	var menuitemsNotActive int // не активные - счетчик

	var menuitemsFoundInWoo int    // найдены в WOO
	var menuitemsNotFoundInWoo int // не найдены в WOO

ForBreak:
	for _, menuitem := range menuitems {
		logger.Debug("--------------------------------------")
		imageNamesRK7 := menuitem.GetImageNames()
		dish := fmt.Sprintf("Блюдо RK7: Name: %s, LongName: %s, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status: %d, RK_Price: %d, RK_Image: %v",
			menuitem.Name, menuitem.WOO_LONGNAME, menuitem.ItemIdent, menuitem.WOO_ID, menuitem.WOO_PARENT_ID, menuitem.Status, menuitem.PRICETYPES, imageNamesRK7)
		logger.Debugf(dish)

		i := ImageSync{IdentRK: menuitem.Ident}

		logger.Debug("Проверка игнор-лист")
		for _, ignoreIdent := range cfg.RK7.MenuitemIdentIgnore {
			if menuitem.ItemIdent == ignoreIdent {
				logger.Debug("Блюдо в игнор-листе")
				i.Status = IMAGE_STATUS_IGNORE
				*imagesSync = append(*imagesSync, i)
				continue ForBreak
			}
		}

		logger.Debug("Проверяем наличие в стоп-листе")
		if dishRests, foundInStopList := dishRestsByIdent[menuitem.Ident]; foundInStopList {
			if dishRests.Prohibited == 1 || dishRests.Quantity == 0 {
				logger.Debug("Блюдо в стоп-листе")
				i.Status = IMAGE_STATUS_IGNORE
				*imagesSync = append(*imagesSync, i)
				continue ForBreak
			}
		}

		if menuitem.Status == 3 {
			logger.Debug("Блюдо активное. Проверяем цену/стоп-лист/включение синхронизации")
			menuitemsActive++
			if menuitem.PRICETYPES != 9223372036854775807 {
				if _, found := productsWooByID[menuitem.WOO_ID]; found {
					logger.Debug("Блюдо найдено в WOO")
					menuitemsFoundInWoo++
					i.IdentWOO = menuitem.WOO_ID
					logger.Debugf("Всего картинок %d штук", len(imageNamesRK7))
					if len(imageNamesRK7) > 0 {
						for _, imageNameRK7 := range imageNamesRK7 {
							if imageFileInDbMap, found := imageFilesInDbMap[imageNameRK7]; found {
								logger.Debug("Найдена картинка в папке и WOO.Media: Name=%s, ID=%d", imageNameRK7, imageFileInDbMap.IdentWOO)
								modtime, err := time.Parse(time.RFC3339, imageFileInDbMap.ModTime)
								if err != nil {
									return errors.Wrapf(err, "failed in time.Parse(%s)", imageFileInDbMap.ModTime)
								}
								i.Images = append(i.Images, Image{
									Name:     imageNameRK7,
									ModTime:  modtime,
									IdentWOO: imageFileInDbMap.IdentWOO,
									IsFound:  true, // картинка найдена
								})
							} else {
								logger.Debug("Картинка не найдена в папке и WOO.Media: Name=%s", imageNameRK7)
								i.Images = append(i.Images, Image{
									Name:    imageNameRK7,
									IsFound: false, // картинка не найдена
								})
							}
						}
						i.Status = IMAGE_STATUS_NEED_VERIFY
						*imagesSync = append(*imagesSync, i)
					} else {
						logger.Debug("Блюдо без картинок. Обнуляем")
						i.Status = IMAGE_STATUS_RK7_IMAGE_NAME_NOT_FOUND
						*imagesSync = append(*imagesSync, i)
					}
				} else {
					logger.Debug("Блюдо без WOO_ID. Игнорировать. Синхронизация меню должна указать WOO")
					menuitemsNotFoundInWoo++
					i.Status = IMAGE_STATUS_RK7_WOO_ID_NOT_FOUND
					*imagesSync = append(*imagesSync, i)
				}
			} else {
				logger.Debug("Блюдо - цена не указана. Игнорируем. Синхронизация меню отключила/удалила блюдо в WOO")
				i.Status = IMAGE_STATUS_IGNORE
				*imagesSync = append(*imagesSync, i)
			}
		} else {
			logger.Debug("Не активное блюдо. Игнорируем. Синхронизация меню отключила/удалила блюдо в WOO")
			menuitemsNotActive++
			i.Status = IMAGE_STATUS_IGNORE
			*imagesSync = append(*imagesSync, i)
		}
	}

	return nil
}
