package categlist

import (
	"WooWithRkeeper/pkg/logging"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func SyncCateglist(db *sqlx.DB) error {
	logger := logging.GetLogger()
	logger.Info("Start SyncCateglist")
	defer logger.Info("End SyncCateglist")

	// 1 этап - синхронизация RK7.Categlist в DB.Categlist и WOO.ProductCategory
	err := HandlerCateglistToDb(db)
	if err != nil {
		return errors.Wrap(err, "failed in HandlerCateglistToDb")
	}

	// 2 этап - обработка DB.Categlist
	err = HandlerCateglistDbOneStage(db)
	if err != nil {
		return errors.Wrap(err, "failed in HandlerCateglistDbOneStage")
	}

	///////////////// todo кажется что временная срань /////////////////
	// TODO надо получить статус NO_NEED_UPDATE - переделать на PARENT_ID_NEED_UPDATE - после создания папки
	// 1 этап - синхронизация RK7.Categlist в DB.Categlist и WOO.ProductCategory
	err = HandlerCateglistToDb(db)
	if err != nil {
		return errors.Wrap(err, "failed in HandlerCateglistToDb")
	}
	///////////////// todo кажется что временная срань /////////////////

	///////////////// todo кажется что временная срань /////////////////
	// todo отрабатывает только на NO_NEED_UPDATE
	// 3 этап - синхронизация DB.Categlist.Parent и WOO.ProductCategory.Parent
	err = HandlerCateglistUpdateParentId(db)
	if err != nil {
		return errors.Wrap(err, "failed in HandlerDbImage")
	}
	///////////////// todo кажется что временная срань /////////////////

	return nil
}
