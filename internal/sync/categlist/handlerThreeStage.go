package categlist

import (
	"WooWithRkeeper/internal/cache"
	"WooWithRkeeper/internal/config"
	"WooWithRkeeper/internal/rk7api"
	"WooWithRkeeper/internal/rk7api/models"
	"WooWithRkeeper/internal/telegram"
	"WooWithRkeeper/internal/wooapi"
	"WooWithRkeeper/pkg/logging"
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

func VerifyAndUpdateParentIdInRk(categlistsSync []CateglistSync) error {
	logger := logging.GetLogger()
	logger.Debug("Start VerifyAndUpdateParentIdInRk")
	defer logger.Debug("End VerifyAndUpdateParentIdInRk")

	cfg := config.GetConfig()
	rk7 := rk7api.GetAPI("REF")

	menu, err := cache.GetMenu()
	if err != nil {
		return errors.Wrap(err, "failed in cache.GetMenu()")
	}

	categlistsRK7ByIdent, err := menu.GetCateglistsRK7ByIdent()
	if err != nil {
		return errors.Wrap(err, "failed in GetCateglistsRK7ByIdent()")
	}

	var m []string
	m = append(m, "<strong>Ошибки при обновлении ParentID у папок RK7</strong>")
	for _, categlistSync := range categlistsSync { // TODO риск не знаю что произойдет
		categlistInRk7 := categlistSync.Categlist
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
				}
			} else {
				logger.Debug("Обновление WOO_PARENT_ID в RK7 не требуется")
			}
		} else {
			m = append(m, fmt.Sprintf("Папка Parent(ID=%d) не найдена в RK7", categlistInRk7.MainParentIdent))
		}
	}
	if len(m) > 1 {
		telegram.SendMessageToTelegramWithLogError(strings.Join(m, "\n"))
	}

	return nil
}

func VerifyAndUpdateParentIdInWoo(categlistsSync []CateglistSync) error {
	logger := logging.GetLogger()
	logger.Debug("Start VerifyAndUpdateParentIdInWoo")
	defer logger.Debug("End VerifyAndUpdateParentIdInWoo")

	var err error
	woo := wooapi.GetAPI()
	if err != nil {
		return errors.Wrap(err, "failed in wooapi.GetAPI()")
	}

	menu, err := cache.GetMenu()
	if err != nil {
		return errors.Wrap(err, "failed in cache.GetMenu()")
	}

	productCategoriesWooByID, err := menu.GetProductCategoriesWooByID()
	if err != nil {
		return errors.Wrap(err, "failed in GetProductCategoriesWooByID()")
	}

	var m []string
	m = append(m, "<strong>Ошибки при обновлении ParentID у папок WOO</strong>")
	for _, categlistSync := range categlistsSync { // TODO риск не знаю что произойдет
		categlistInRk7 := categlistSync.Categlist
		if categlistInRk7.WOO_ID != 0 {
			if categlistInRk7.WOO_PARENT_ID != 0 {
				if productCategory, found := productCategoriesWooByID[categlistInRk7.WOO_ID]; found {
					logger.Debugf("ProductCategory успешно найден: %s", GetProductCategoryDescription(productCategory))
					if categlistInRk7.WOO_PARENT_ID != productCategory.Parent {
						recoveryParent := productCategory.Parent
						productCategory.Parent = categlistInRk7.WOO_PARENT_ID
						_, err = woo.ProductCategoryUpdate(productCategory)
						if err != nil {
							productCategory.Parent = recoveryParent
							errorText := fmt.Sprintf("failed in ProductCategoryUpdate(ID=%d, Name=%s, Parent=%d); %v", productCategory.ID, productCategory.Name, categlistInRk7.WOO_PARENT_ID, err)
							m = append(m, errorText)
						} else {
							logger.Debug("Папка успешно обновлена. Кеш обновлен")
						}
					} else {
						logger.Debug("Обновление не требуется")
					}
				} else {
					m = append(m, fmt.Sprintf("%s; Не найден в WOO", GetCateglistDescription(categlistInRk7)))
				}
			} else {
				m = append(m, fmt.Sprintf("%s; WOO_PARENT_ID = 0", GetCateglistDescription(categlistInRk7)))
			}
		} else {
			m = append(m, fmt.Sprintf("%s; WOO_ID = 0", GetCateglistDescription(categlistInRk7)))
		}
	}

	if len(m) > 1 {
		telegram.SendMessageToTelegramWithLogError(strings.Join(m, "\n"))
	}

	return nil
}

func HandlerNeedUpdateParentId(categlistsSync *[]CateglistSync) error {
	logger := logging.GetLogger()
	logger.Debug("Start HandlerNeedUpdate")
	defer logger.Debug("End HandlerNeedUpdate")

	categlistsSyncByStatus := make(map[string][]CateglistSync)
	for _, categlistSync := range *categlistsSync {
		categlistsSyncByStatus[categlistSync.StatusSync] = append(categlistsSyncByStatus[categlistSync.StatusSync], categlistSync)
	}

	for status, categlistsSync := range categlistsSyncByStatus {
		if status == NOT_WOO_ID || status == NOT_FOUND_IN_WOO || status == NEED_UPDATE || status == NOT_NEED_UPDATE {
			err := VerifyAndUpdateParentIdInRk(categlistsSync)
			if err != nil {
				return errors.Wrap(err, "failed in VerifyAndUpdateParentIdInRk")
			}
		}
	}

	for status, categlistsSync := range categlistsSyncByStatus {
		if status == NOT_WOO_ID || status == NOT_FOUND_IN_WOO || status == NEED_UPDATE || status == NOT_NEED_UPDATE {
			err := VerifyAndUpdateParentIdInWoo(categlistsSync)
			if err != nil {
				return errors.Wrap(err, "failed in VerifyAndUpdateParentIdInWoo")
			}
		}
	}

	return nil
}
