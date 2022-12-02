package model

import (
	"WooWithRkeeper/pkg/logging"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

const (
	IMAGE_STATUS_IGNORE                           = "Ignore"                     // Игнор
	IMAGE_STATUS_WOO_NOT_FOUND                    = "WooNotFound"                // Product не найден в WOO
	IMAGE_STATUS_RK7_WOO_ID_NOT_FOUND             = "Rk7WooIDNotFound"           // Не указан WOO_ID в RK7
	IMAGE_STATUS_RK7_IMAGE_NAME_NOT_FOUND         = "Rk7ImageNameNotFound"       // Удаляем в WOO - Не указан IMAGE_NAME в RK7
	IMAGE_STATUS_FILE_NOT_FOUND                   = "FileNotFound"               // Удаляем в WOO - Не найдена файл картинки
	IMAGE_STATUS_NEED_UPDATE_BY_DIFF_NAME         = "NeedUpdateByDiffName"       // Обновляем - имя картинка изменилась
	IMAGE_STATUS_NEED_UPDATE_BY_DIFF_DATE         = "NeedUpdateByDiffDate"       // Обновляем - дата картинки изменилась
	IMAGE_STATUS_NEED_UPDATE_BY_NOT_FOUND_IN_DB   = "NeedUpdateByNotFoundInDb"   // Сообщаем об ошибке
	IMAGE_STATUS_NEED_UPDATE_BY_FIND_DOUBLE_IN_DB = "NeedUpdateByFindDoubleInDb" // Сообщаем об ошибке
	IMAGE_STATUS_NO_NEED_UPDATE                   = "NoNeedUpdate"               // Проверить, что все ок

)

type Image struct {
	ID       int            `db:"ID"`
	IdentRK  int            `db:"IdentRK"`
	ModTime  sql.NullString `db:"ModTime"`  //todo перепроверить переменные
	Name     sql.NullString `db:"Name"`     //todo перепроверить переменные
	Pos      sql.NullInt32  `db:"Pos"`      //todo проверить это 0 или null, когда не указан
	IdentWOO sql.NullInt32  `db:"IdentWOO"` //todo проверить это 0 или null, когда не указан
	Status   sql.NullString `db:"Status"`   //todo Status https://docs.google.com/spreadsheets/d/1oZ7jDxDHMfHvsLfN90HYmV3Cdj6uO_hS4MVBGJzuKqU/edit#gid=1756609729
}

func (i *Image) SelectByStatus(db *sqlx.DB) ([]*Image, error) {
	logger := logging.GetLogger()
	logger.Debug("Start Image.SelectByIdentRKAndPos")
	defer logger.Debug("End Image.SelectByIdentRKAndPos")

	var err error
	var imagesInDb []*Image
	var query string

	logger.Debug("Выполняем поиск записей в таблице Images")
	if i.Status.Valid {
		query = "SELECT * FROM Image WHERE Status=$1;"
		err = db.Select(&imagesInDb, query, i.Status)
		logger.Debugf("SELECT:\n%s(%s)", query, i.Status.String)
	} else {
		logger.Debugf("SELECT:\n%s(%v)", query, i.Status)
		return nil, errors.New("Неизвестная ошибка, status is null")
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed SELECT to dbsqlite; query:\n%s(%d, %v)", query, i.IdentRK, i.Pos)
	}

	logger.Debugf("Количество полученных строк: %d", len(imagesInDb))
	return imagesInDb, nil

}

func (i *Image) SelectByIdentRKAndPos(db *sqlx.DB) ([]*Image, error) {
	logger := logging.GetLogger()
	logger.Debug("Start Image.SelectByIdentRKAndPos")
	defer logger.Debug("End Image.SelectByIdentRKAndPos")

	var err error
	var imagesInDb []*Image
	var query string

	logger.Debug("Выполняем поиск записей в таблице Images")
	if i.Pos.Valid {
		query = "SELECT * FROM Image WHERE IdentRK=$1 AND Pos=$2;"
		err = db.Select(&imagesInDb, query, i.IdentRK, i.Pos)
		logger.Debugf("SELECT:\n%s(%d, %v)", query, i.IdentRK, i.Pos)
	} else {
		query = "SELECT * FROM Image WHERE IdentRK=$1;"
		err = db.Select(&imagesInDb, query, i.IdentRK)
		logger.Debugf("SELECT:\n%s(%d)", query, i.IdentRK)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed SELECT to dbsqlite; query:\n%s(%d, %v)", query, i.IdentRK, i.Pos)
	}

	logger.Debugf("Количество полученных строк: %d", len(imagesInDb))
	return imagesInDb, nil

}

func (i *Image) UpdateByIdentRKAndPos(db *sqlx.DB) error {

	logger := logging.GetLogger()
	logger.Debug("Start Image.UpdateByIdentRKAndPos")
	defer logger.Debug("End Image.UpdateByIdentRKAndPos")

	var err error
	var query string

	imagesInDb, err := i.SelectByIdentRKAndPos(db)
	if err != nil {
		return errors.Wrapf(err, "failed in SelectByIdentRKAndPos()")
	}

	logger.Debugf("Количество найденных строк: %d", len(imagesInDb))
	switch {
	case len(imagesInDb) == 0:
		logger.Debug("Строка не найдена, требуется ее добавить")
		tx := db.MustBegin()
		query = "INSERT INTO Image (IdentRK, ModTime, Name, Pos, IdentWOO, Status) VALUES ($1, $2, $3, $4, $5, $6);"
		logger.Debugf("INSERT:\n%s(%v)", query, i)
		tx.MustExec(query, i.IdentRK, i.ModTime, i.Name, i.Pos, i.IdentWOO, i.Status)
		err := tx.Commit()
		if err != nil {
			return errors.Wrapf(err, "failed INSERT to dbsqlite; query:\n%s(%v)", query, i)
		} else {
			logger.Debug("Строка добавлена успешно")
			return nil
		}
	case len(imagesInDb) == 1:
		if imagesInDb[0].ModTime != i.ModTime ||
			imagesInDb[0].Name != i.Name || imagesInDb[0].IdentWOO != i.IdentWOO ||
			imagesInDb[0].Status != i.Status {
			logger.Debug("Требуется обновление строки")
			logger.Debug(imagesInDb[0].ModTime, i.ModTime)
			logger.Debug(imagesInDb[0].Name, i.Name)
			logger.Debug(imagesInDb[0].IdentWOO, i.IdentWOO)
			logger.Debug(imagesInDb[0].Status, i.Status)
			if i.Pos.Valid {
				query = "UPDATE Image SET ModTime=:ModTime, Name=:Name, IdentWOO=:IdentWOO, Status=:Status WHERE IdentRK=:IdentRK AND Pos=:Pos;"
			} else {
				query = "UPDATE Image SET ModTime=:ModTime, Name=:Name, IdentWOO=:IdentWOO, Status=:Status WHERE IdentRK=:IdentRK;"
			}
			logger.Debugf("UPDATE:\n%s(%v)", query, i)
			_, err = db.NamedExec(query,
				map[string]interface{}{
					"IdentRK":  i.IdentRK,
					"ModTime":  i.ModTime,
					"Name":     i.Name,
					"Pos":      i.Pos,
					"IdentWOO": i.IdentWOO,
					"Status":   i.Status,
				})
			if err != nil {
				return errors.Wrapf(err, "failed UPDATE to dbsqlite; query:\n%s(%v)", query, i)
			} else {
				logger.Debug("Строка обновлена успешно")
				return nil
			}
		} else {
			logger.Debug("Обновление строки не требуется")
			return nil
		}
	case len(imagesInDb) > 1:
		logger.Debug("Необходимо удалить дублирующие строки с полями IdentRK+Pos")
		tx := db.MustBegin()
		if i.Pos.Valid {
			query = "DELETE FROM Image WHERE IdentRK=$1 AND Pos=$2;"
			tx.MustExec(query, i.IdentRK, i.Pos)
			logger.Debugf("DELETE:\n%s(%d)", query, i.IdentRK, i.Pos)
		} else {
			query = "DELETE FROM Image WHERE IdentRK=$1;"
			tx.MustExec(query, i.IdentRK)
			logger.Debugf("DELETE:\n%s(%d)", query, i.IdentRK)
		}
		err = tx.Commit()
		if err != nil {
			return errors.Wrapf(err, "failed DELETE in dbsqlite; query:\n%s(%d)", query, i.IdentRK)
		} else {
			logger.Debug("Строки удалены успешно")
			logger.Debug("Требуется добавить строку")
			tx := db.MustBegin()
			query = "INSERT INTO Image (IdentRK, ModTime, Name, Pos, IdentWOO, Status) VALUES ($1, $2, $3, $4, $5, $6);"
			logger.Debugf("INSERT:\n%s(%v)", query, i)
			tx.MustExec(query, i.IdentRK, i.ModTime, i.Name, i.Pos, i.IdentWOO, i.Status)
			err := tx.Commit()
			if err != nil {
				return errors.Wrapf(err, "failed INSERT to dbsqlite; query:\n%s(%v)", query, i)
			} else {
				logger.Debug("Строка добавлена успешно")
				return nil
			}
		}
	default:
		return errors.New("Неизвестная ошибка")
	}
}
