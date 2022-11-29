package database

import (
	"WooWithRkeeper/pkg/logging"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Version struct {
	ID      int    `db:"ID"`
	Name    string `db:"Name"`
	Version int    `db:"Version"`
}

type Image struct {
	ID             int            `db:"ID"`
	IdentRK        int            `db:"IdentRK"`
	IMAGE_MOD_TIME sql.NullString `db:"ImageModTime"` //todo перепроверить переменные
	IMAGE_NAME     sql.NullString `db:"ImageName"`    //todo перепроверить переменные
	Pos            sql.NullInt32  `db:"Pos"`          //todo проверить это 0 или null, когда не указан
	IdentWOO       sql.NullInt32  `db:"IdentWOO"`     //todo проверить это 0 или null, когда не указан
	Status         sql.NullString `db:"Status"`       //todo Status https://docs.google.com/spreadsheets/d/1oZ7jDxDHMfHvsLfN90HYmV3Cdj6uO_hS4MVBGJzuKqU/edit#gid=1756609729
}

const DATABASE_SELECT_MENUITEM = `SELECT ID, IdentRK, ImageModTime1, ImageName1, ImageModTime2, ImageName2 FROM Menuitem WHERE IdentRK=%d`

func (i *Image) UpdateInDb(db *sqlx.DB) error {

	logger := logging.GetLogger()
	logger.Debug("Start Image.UpdateInDb")
	defer logger.Debug("End Image.UpdateInDb")

	logger.Debug("Выполняем поиск записей в таблице Images")
	var imagesInDb []Image
	query := "SELECT ID, IdentRK, ImageModTime, ImageName, Pos, IdentWOO, Status FROM Image WHERE IdentRK=$1;"
	err := db.Select(&imagesInDb, query, i.IdentRK)
	if err != nil {
		return errors.Wrapf(err, "failed SELECT to dbsqlite; query:\n%s(%d)", query, i.IdentRK)
	} else {
		logger.Debugf("Количество найденных строк: %d", len(imagesInDb))
		switch len(imagesInDb) {
		case 0:
			tx := db.MustBegin()
			query := "INSERT INTO Image (IdentRK, ImageModTime, ImageName, Pos, IdentWOO, Status) VALUES ($1, $2, $3, $4, $5, $6);"
			logger.Debugf("INSERT:\n%s(%d, %v, %v, %v, %v, %v)", query, i.IdentRK, i.IMAGE_MOD_TIME, i.IMAGE_NAME, i.Pos, i.IdentWOO, i.Status)
			tx.MustExec(query, i.IdentRK, i.IMAGE_MOD_TIME, i.IMAGE_NAME, i.Pos, i.IdentWOO, i.Status)
			err := tx.Commit()
			if err != nil {
				return errors.Wrapf(err, "failed INSERT to dbsqlite; query:\n%s(%d, %v, %v, %v, %v, %v)",
					query, i.IdentRK, i.IMAGE_MOD_TIME, i.IMAGE_NAME, i.Pos, i.IdentWOO, i.Status)
			} else {
				return nil
			}
		case 1:
			query := `UPDATE Image SET ImageModTime=:IMAGE_MOD_TIME, ImageName=:IMAGE_NAME, Pos=:Pos, IdentWOO=:IdentWOO, Status=:Status WHERE IdentRK=:IdentRK;`
			logger.Debugf("UPDATE:\n%s(%v)", query, i)
			_, err = db.NamedExec(query,
				map[string]interface{}{
					"IdentRK":        i.IdentRK,
					"IMAGE_MOD_TIME": i.IMAGE_MOD_TIME,
					"IMAGE_NAME":     i.IMAGE_NAME,
					"Pos":            i.Pos,
					"IdentWOO":       i.IdentWOO,
					"Status":         i.Status,
				})
			if err != nil {
				return errors.Wrapf(err, "failed UPDATE to dbsqlite; query:\n%s(%d, %v, %v, %v, %v, %v)",
					query, i.IdentRK, i.IMAGE_MOD_TIME, i.IMAGE_NAME, i.Pos, i.IdentWOO, i.Status)
			} else {
				return nil
			}
		default:
			tx := db.MustBegin()
			query := "DELETE FROM Image WHERE IdentRK=$1;"
			logger.Debugf("DELETE:\n%s(%d)", query, i.IdentRK)
			tx.MustExec(query, i.IdentRK)
			err = tx.Commit()
			if err != nil {
				return errors.Wrapf(err, "failed DELETE in dbsqlite; query:\n%s(%d)", query, i.IdentRK)
			} else {
				tx := db.MustBegin()
				query := "INSERT INTO Image (IdentRK, ImageModTime, ImageName, Pos, IdentWOO, Status) VALUES ($1, $2, $3, $4, $5, $6);"
				logger.Debugf("INSERT:\n%s(%d, %v, %v, %v, %v, %v)", query, i.IdentRK, i.IMAGE_MOD_TIME, i.IMAGE_NAME, i.Pos, i.IdentWOO, i.Status)
				tx.MustExec(query, i.IdentRK, i.IMAGE_MOD_TIME, i.IMAGE_NAME, i.Pos, i.IdentWOO, i.Status)
				err := tx.Commit()
				if err != nil {
					return errors.Wrapf(err, "failed INSERT to dbsqlite; query:\n%s(%d, %v, %v, %v, %v, %v)",
						query, i.IdentRK, i.IMAGE_MOD_TIME, i.IMAGE_NAME, i.Pos, i.IdentWOO, i.Status)
				} else {
					return nil
				}
			}
		}
	}
}

const (
	IMAGE_STATUS_IGNORE                           = "Ignore"
	IMAGE_STATUS_IN_WOO_NOT_FOUND                 = "InWooNotFound"
	IMAGE_STATUS_RK7_WOO_ID_NOT_FOUND             = "Rk7WooIDNotFound"
	IMAGE_STATUS_RK7_IMAGE_NAME_NOT_FOUND         = "Rk7ImageNameNotFound"
	IMAGE_STATUS_IMAGE_FILE_NOT_FOUND             = "ImageFileNotFound"
	IMAGE_STATUS_IMAGE_FILE_NOT_IS_JPEG           = "ImageFileNotIsJpg"
	IMAGE_STATUS_NEED_UPDATE_BY_DIFF_NAME         = "NeedUpdateByDiffName"
	IMAGE_STATUS_NEED_UPDATE_BY_DIFF_DATE         = "NeedUpdateByDiffDate"
	IMAGE_STATUS_NEED_UPDATE_BY_NOT_FOUND_IN_DB   = "NeedUpdateByNotFoundInDb"
	IMAGE_STATUS_NEED_UPDATE_BY_FIND_DOUBLE_IN_DB = "NeedUpdateByFindDoubleInDb"
	IMAGE_STATUS_NO_NEED_UPDATE                   = "NoNeedUpdate"
)

const DB_SCHEMA = `CREATE TABLE Version (
	ID integer PRIMARY KEY AUTOINCREMENT,
	Name text,
	Version integer
);

CREATE TABLE Image (
	ID integer PRIMARY KEY AUTOINCREMENT,
	IdentRK integer,
	ImageModTime text,
	ImageName text,
	Pos integer, 
	IdentWOO integer,
	Status text
);
`
