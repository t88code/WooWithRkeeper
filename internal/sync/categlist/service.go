package categlist

import (
	"WooWithRkeeper/internal/rk7api/models"
	modelsWOOAPI "WooWithRkeeper/internal/wooapi/models"
	"WooWithRkeeper/pkg/logging"
	"fmt"
	"github.com/pkg/errors"
)

func SyncCateglist() error {
	logger := logging.GetLogger()
	logger.Info("Start SyncCateglist")
	defer logger.Info("End SyncCateglist")

	categlistsSync := make([]CateglistSync, 0)

	logger.Debug("Запускаем 1 этап синхронизации папок")
	// 1 этап - синхронизация RK7.Categlist в DB.Categlist и WOO.ProductCategory
	err := HandlerCateglistToDb(&categlistsSync)
	if err != nil {
		return errors.Wrap(err, "failed in HandlerCateglistToDb")
	} else {
		logger.Debug("1 этап синхронизации папок успешно завершен")
	}

	logger.Debug("Запускаем 2 этап синхронизации папок")
	// 2 этап - обработка DB.Categlist
	err = HandlerCateglistDbOneStage(&categlistsSync)
	if err != nil {
		return errors.Wrap(err, "failed in HandlerCateglistDbOneStage")
	} else {
		logger.Debug("2 этап синхронизации папок успешно завершен")
	}

	logger.Debug("Запускаем 3 этап синхронизации папок")
	// 3 этап - синхронизация DB.Categlist.Parent и WOO.ProductCategory.Parent
	err = HandlerNeedUpdateParentId(&categlistsSync)
	if err != nil {
		return errors.Wrap(err, "failed in HandlerNeedUpdateParentId")
	} else {
		logger.Debug("3 этап синхронизации папок успешно завершен")
	}

	return nil
}

// GetCateglistDescription Сформировать строку с папкой RK7
// Используется в 1, 2 этапе синхронизации папок
func GetCateglistDescription(categlist *models.Categlist) string {
	return fmt.Sprintf("Name: %s, Longname: %s, RK_CODE: %d, RK_ID: %d, RK_WOO_ID: %d, RK_WOO_PARENT_ID: %d, RK7_Status %d, WOO_SYNC %d", categlist.Name, categlist.WOO_LONGNAME, categlist.Code, categlist.ItemIdent, categlist.WOO_ID, categlist.WOO_PARENT_ID, categlist.Status, categlist.WOO_SYNC)
}

// GetProductCategoryDescription Сформировать строку с папкой WOO
// Используется в 2 этапе синхронизации папок
func GetProductCategoryDescription(productCategory *modelsWOOAPI.ProductCategory) string {
	return fmt.Sprintf("Name: %s, ID: %d, Parent: %d", productCategory.Name, productCategory.ID, productCategory.Parent)
}
