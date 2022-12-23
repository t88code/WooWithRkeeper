package menuitem

import (
	"WooWithRkeeper/pkg/logging"
	"github.com/pkg/errors"
)

func SyncMenuitems() error {
	logger := logging.GetLogger()
	logger.Info("Start SyncMenuitems")
	defer logger.Info("End SyncMenuitems")

	var err error
	menuitemsSync := make([]MenuitemSync, 0)

	// HandlerOneStage - 1 стадия синхронизации, сверка между RK7/WOO
	err = HandlerOneStage(&menuitemsSync)
	if err != nil {
		return errors.Wrap(err, "failed in HandlerOneStage()")
	}

	// HandlerTwoStage - 2 стадия синхронизации, запуск обработчиков по каждому статусу
	err = HandlerTwoStage(&menuitemsSync)
	if err != nil {
		return errors.Wrap(err, "failed in HandlerTwoStage()")
	}

	return nil
}
