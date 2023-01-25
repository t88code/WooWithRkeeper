package images

import (
	"WooWithRkeeper/internal/cache"
	"WooWithRkeeper/internal/telegram"
	"WooWithRkeeper/internal/wooapi"
	modelsWOOAPI "WooWithRkeeper/internal/wooapi/models"
	"WooWithRkeeper/pkg/logging"
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

// HandlerDbImage 3 этап - Синхронизация DB.Image/DB.ImageFiles с Woo.Product.Image
func HandlerDbImage(imagesSync *[]ImageSync) error {

	logger := logging.GetLogger()
	logger.Debug("Start HandlerDbImage")
	defer logger.Debug("End HandlerDbImage")
	var err error

	imagesSyncByStatus := make(map[string][]ImageSync)
	for _, imageSync := range *imagesSync {
		imagesSyncByStatus[imageSync.Status] = append(imagesSyncByStatus[imageSync.Status], imageSync)
	}

	err = HandlerNulledImage(imagesSyncByStatus[IMAGE_STATUS_RK7_IMAGE_NAME_NOT_FOUND])
	if err != nil {
		return errors.Wrap(err, "failed in HandlerNulledImage")
	}

	err = HandlerImageVerify(imagesSyncByStatus[IMAGE_STATUS_NEED_VERIFY])
	if err != nil {
		return errors.Wrap(err, "failed in HandlerImageVerify")
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Обнуление картинки
// Используется в 3 этапе синхронизации картинок
func HandlerNulledImage(imagesSync []ImageSync) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerNulledImage")
	defer logger.Debug("End HandlerNulledImage")

	m := []string{"<strong>Не удалось обнулить картинки в WOO</strong>"}
	for _, imageSync := range imagesSync {
		if imageSync.Status == IMAGE_STATUS_RK7_IMAGE_NAME_NOT_FOUND {
			err := NulledImageInWoo(imageSync.IdentWOO)
			if err != nil {
				m = append(m, err.Error())
			}
		}
	}
	if len(m) > 1 {
		telegram.SendMessageToTelegramWithLogError(strings.Join(m, "\n"))
	}
	return nil
}

// Обнуление картинки
// Используется в 3 этапе синхронизации картинок
func NulledImageInWoo(wooID int) error {
	logger := logging.GetLogger()
	logger.Debug("Start NulledImageInWoo")
	defer logger.Debug("End NulledImageInWoo")

	logger.Debugf("Обнуляем картинки в WOO(id=%d)", wooID)

	menu, err := cache.GetMenu()
	if err != nil {
		return errors.Wrap(err, "failed in cache.GetMenu()")
	}

	productsWooByID, err := menu.GetProductsWooByID()
	if err != nil {
		return errors.Wrap(err, "failed in menu.GetProductsWooByID()")
	}
	apiWoo := wooapi.GetAPI()

	if _, found := productsWooByID[wooID]; found {
		recoveryImages := productsWooByID[wooID].Images
		productsWooByID[wooID].Images = make([]modelsWOOAPI.ProductImage, 0)
		_, err := apiWoo.ProductUpdate(productsWooByID[wooID])
		if err != nil {
			productsWooByID[wooID].Images = recoveryImages
			return errors.Wrapf(err, "Не удалось обнулить WOO(id=%d)", wooID)
		} else {
			logger.Debug("Картинки обнулены")
		}
	} else {
		return errors.New(fmt.Sprintf("Не удалось найти в WOO(id=%d)", wooID))
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Обновление картинок
// Используется в 3 этапе синхронизации картинок
func HandlerImageVerify(imagesSync []ImageSync) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerImageVerify")
	defer logger.Debug("End HandlerImageVerify")

	menu, err := cache.GetMenu()
	if err != nil {
		return errors.Wrap(err, "failed in cache.GetMenu()")
	}

	productsWooByID, err := menu.GetProductsWooByID()
	if err != nil {
		return errors.Wrap(err, "failed in menu.GetProductsWooByID()")
	}

	m := []string{"<strong>Ошибки при обновлении картинок между WOO/RK7</strong>"}
	for _, imageSync := range imagesSync {
		if product, found := productsWooByID[imageSync.IdentWOO]; found {
			// проверка картинок №1 - сверяем все позиции
			index := 0
			needUpdate := false
			for _, imageSync := range imageSync.Images {
				if imageSync.IsFound {
					if len(product.Images) > index {
						if product.Images[index].Id != imageSync.IdentWOO {
							logger.Debug("Картинка не совпадает между WOO/RK7. Необходимо обновить")
							needUpdate = true
							break
						} else {
							logger.Debug("Картинка совпадает между WOO/RK7")
							index++
						}
					} else {
						logger.Debugf("Количество картинок WOO=%d меньше RK7=%d. Требуется обновление", len(product.Images), index+1)
						needUpdate = true
						break
					}
				}
			}
			if (!needUpdate && len(product.Images) != index) || needUpdate {
				logger.Debug("Требуется обновление картинок")
				err := UpdateImagesInWoo(imageSync)
				if err != nil {
					m = append(m, fmt.Sprintf("Не удалось обновить продукт в WOO: WOOID=%d", imageSync.IdentWOO))
				}
			}
		} else {
			m = append(m, fmt.Sprintf("Продукт не найден: WOOID=%d", imageSync.IdentWOO))
		}

	}
	if len(m) > 1 {
		telegram.SendMessageToTelegramWithLogError(strings.Join(m, "\n"))
	}

	m = []string{"<strong>Картинка не найдена в папке Images</strong>"}
	for _, imageSync := range imagesSync {
		// проверка картинок №2 - находим картинки которых нет в папке
		for _, image := range imageSync.Images {
			if !image.IsFound && image.Name != "" {
				logger.Debug("Картинка не найдена. Отправить сообщение")
				m = append(m, fmt.Sprintf("Name=%s, Блюдо: WOO_ID=%d, RK_ID=%d", image.Name, imageSync.IdentWOO, imageSync.IdentRK))
			}
		}

	}
	if len(m) > 1 {
		telegram.SendMessageToTelegramWithLogError(strings.Join(m, "\n"))
	}

	return nil
}

// Обновление картинок
// Используется в 3 этапе синхронизации картинок
// update // TODO !проверить!!!!
func UpdateImagesInWoo(imageSync ImageSync) error {
	logger := logging.GetLogger()
	logger.Debug("Start UpdateImagesInWoo")
	defer logger.Debug("End UpdateImagesInWoo")

	menu, err := cache.GetMenu()
	if err != nil {
		return errors.Wrap(err, "failed in cache.GetMenu()")
	}
	menuitemsRK7ByIdent, err := menu.GetMenuitemsRK7ByIdent()
	if err != nil {
		return errors.Wrap(err, "failed in menu.GetMenuitemsRK7ByIdent()")
	}
	productsWooByID, err := menu.GetProductsWooByID()
	if err != nil {
		return errors.Wrap(err, "failed in menu.GetProductsWooByID()")
	}

	apiWoo := wooapi.GetAPI()

	if menuitem, ok := menuitemsRK7ByIdent[imageSync.IdentRK]; ok {
		if menuitem.WOO_ID != 0 {
			if product, ok := productsWooByID[menuitem.WOO_ID]; ok {
				//выполнить update
				logger.Debugf("Добавляем картинку для продукта: ID:%d, Name:%s", product.ID, product.Name)
				recoveryImages := productsWooByID[menuitem.WOO_ID].Images
				productsWooByID[menuitem.WOO_ID].Images = make([]modelsWOOAPI.ProductImage, 0)
				for _, imageAdd := range imageSync.Images {
					logger.Debugf("Картинка: ID:%d, Name:%s, IdentRK: %v", imageAdd.IdentWOO, imageAdd.Name, imageAdd.ModTime)
					productsWooByID[menuitem.WOO_ID].Images = append(productsWooByID[menuitem.WOO_ID].Images,
						modelsWOOAPI.ProductImage{Id: imageAdd.IdentWOO})
				}
				_, err := apiWoo.ProductUpdate(productsWooByID[menuitem.WOO_ID])
				if err != nil {
					productsWooByID[menuitem.WOO_ID].Images = recoveryImages
					return errors.Wrapf(err, fmt.Sprintf("Не удалось обновить product.images: %v", imageSync.Images))
				} else {
					logger.Debug("Картинка обновлена")
					return nil
				}
			} else {
				return errors.New(fmt.Sprintf("Блюдо(id=%d, wooid=%d) не найдено в WOO", imageSync.IdentRK, menuitem.WOO_ID))
			}
		} else {
			return errors.New(fmt.Sprintf("Блюдо(id=%d) без WOO_ID", imageSync.IdentRK))
		}
	} else {
		return errors.New(fmt.Sprintf("Блюдо(id=%d) не найдено в RK7", imageSync.IdentRK))
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
