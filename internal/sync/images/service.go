package images

import (
	"WooWithRkeeper/pkg/logging"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// SyncImages синхронизация картинок
// HandlerImageFileToDb Синхронизация файлов картинок из папки с Woo.Media/DB
// HandlerMenuitemsToDbImage Синхронизация Menuitems c DB.Image
// HandlerDbImage Синхронизация DB.Image/DB.ImageFiles с Woo.Product.Image
func SyncImages(db *sqlx.DB) error {
	logger := logging.GetLogger()
	logger.Debug("Start SyncImages")
	defer logger.Debug("End SyncImages")
	var err error

	// 1 этап - Синхронизация файлов картинок из папки с Woo.Media/DB.ImageFiles
	err = HandlerImageFileToDb(db)
	if err != nil {
		return errors.Wrap(err, "failed in HandlerImageFileToDb")
	}

	// 2 этап - Синхронизация Menuitems c DB.Image
	err = HandlerMenuitemsToDbImage(db)
	if err != nil {
		return errors.Wrap(err, "failed in HandlerMenuitemsToDbImage")
	}

	// 3 этап - Синхронизация DB.Image/DB.ImageFiles с Woo.Product.Image
	err = HandlerDbImage(db)
	if err != nil {
		return errors.Wrap(err, "failed in HandlerDbImage")
	}

	return nil
}
