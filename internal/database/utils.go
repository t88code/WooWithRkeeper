package database

import (
	"WooWithRkeeper/pkg/logging"
	"github.com/jmoiron/sqlx"
	"os"
)

func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func CreateDB(dbname string) error {

	logger := logging.GetLogger()
	logger.Info("CreateDB:>Start")
	defer logger.Info("CreateDB:>End")

	logger.Info("CreateDB:>Creating ", dbname)

	db, err := sqlx.Open("sqlite3", dbname)
	if err != nil {
		logger.Fatal(err)
		return err
	}
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			logger.Error(err)
		}
	}(db)

	logger.Info(dbname, " created")

	db.MustExec(DB_SCHEMA)
	return nil
}
