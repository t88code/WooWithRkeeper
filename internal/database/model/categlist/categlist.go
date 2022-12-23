package categlist

import (
	"WooWithRkeeper/pkg/logging"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

const (
	IGNORE           = "IGNORE"           // Игнор-лист - Обнулить WOO/RK7
	SYNC_OFF         = "SYNC_OFF"         // Синхронизация отключена - Обнуляем в WOO/RK7
	NOT_ACTIVE       = "NOT_ACTIVE"       // Папка не активная - Обнуляем WOO/RK7
	NOT_WOO_ID       = "NOT_WOO_ID"       // Не указан WOO_ID - Создаем в WOO
	NOT_FOUND_IN_WOO = "NOT_FOUND_IN_WOO" // Папка не найдена в WOO - Создаем в WOO
	NEED_UPDATE      = "NEED_UPDATE"      // Папка RK7 не совпадает с WOO(свойства Name/LongName/WOO_ID) - Обновляем WOO
	NOT_NEED_UPDATE  = "NOT_NEED_UPDATE"  // Папка RK7 совпадает с WOO(свойства Name/LongName/WOO_ID) - Обновление в WOO не требуется
)

type Categlist struct {
	ID        int            `db:"ID"`
	Name      sql.NullString `db:"Name"`
	LongName  sql.NullString `db:"LongName"`
	IdentRK   int            `db:"IdentRK"`
	IdentWOO  sql.NullInt32  `db:"IdentWOO"`
	ParentRK  sql.NullInt32  `db:"ParentRK"`
	ParentWOO sql.NullInt32  `db:"ParentWOO"`
	Sync      sql.NullInt32  `db:"Sync"`
	Status    sql.NullString `db:"Status"`
}

func (с *Categlist) SelectByIdentRK(db *sqlx.DB) ([]*Categlist, error) {
	logger := logging.GetLogger()
	logger.Debug("Start Categlist.SelectByIdentRK")
	defer logger.Debug("End Categlist.SelectByIdentRK")

	var err error
	var сateglistInDb []*Categlist
	var query string

	logger.Debug("Выполняем поиск записей в таблице Categlist")
	query = "SELECT * FROM Categlist WHERE IdentRK=$1;"
	err = db.Select(&сateglistInDb, query, с.IdentRK)
	logger.Debugf("SELECT:\n%s(%d)", query, с.IdentRK)
	if err != nil {
		return nil, errors.Wrapf(err, "failed SELECT to dbsqlite; query:\n%s(%d)", query, с.IdentRK)
	}

	logger.Debugf("Количество полученных строк: %d", len(сateglistInDb))
	return сateglistInDb, nil
}

func (c *Categlist) UpdateByIdentRK(db *sqlx.DB) error {

	logger := logging.GetLogger()
	logger.Debug("Start Categlist.UpdateByIdentRK")
	defer logger.Debug("End Categlist.UpdateByIdentRK")

	var err error
	var query string

	сateglistInDb, err := c.SelectByIdentRK(db)
	if err != nil {
		return errors.Wrapf(err, "failed in SelectByIdentRK()")
	}

	logger.Debugf("Количество найденных строк: %d", len(сateglistInDb))

	tx := db.MustBegin()
	defer func() {
		if err != nil {
			logger.Error(err)
			err := tx.Rollback()
			if err != nil {
				logger.Errorf("failed in Rollback(); %v", err)
				return
			} else {
				logger.Info("Rollback() is done")
			}
		}
	}()

	switch {
	case len(сateglistInDb) == 0:
		logger.Debug("Строка не найдена, требуется ее добавить")
		query = `INSERT INTO Categlist (Name, LongName, IdentRK, IdentWOO, ParentRK, ParentWOO, Sync, Status) 
			VALUES (:Name, :LongName, :IdentRK, :IdentWOO, :ParentRK, :ParentWOO, :Sync, :Status);`
		logger.Debugf("INSERT:\n%s(%v)", query, c)
		result, err := tx.NamedExec(query, c)
		if err != nil {
			return errors.Wrapf(err, "failed INSERT to dbsqlite; query:\n%s(%v)", query, c)
		} else {
			l, _ := result.RowsAffected()
			logger.Debugf("Строк %d добавлено успешно", l)
			err := tx.Commit()
			if err != nil {
				return errors.Wrapf(err, "failed in Commit(); failed INSERT to dbsqlite; query:\n%s(%v)", query, c)
			} else {
				logger.Info("Commit() is done")
				return nil
			}
		}
	case len(сateglistInDb) == 1:
		if сateglistInDb[0].Name != c.Name ||
			сateglistInDb[0].LongName != c.LongName ||
			сateglistInDb[0].IdentRK != c.IdentRK ||
			сateglistInDb[0].IdentWOO != c.IdentWOO ||
			сateglistInDb[0].ParentRK != c.ParentRK ||
			сateglistInDb[0].ParentWOO != c.ParentWOO ||
			сateglistInDb[0].Sync != c.Sync ||
			сateglistInDb[0].Status != c.Status {
			logger.Debug("Требуется обновление строки")
			logger.Debug(сateglistInDb[0])
			query = `UPDATE Categlist SET Name=:Name, LongName=:LongName, IdentRK=:IdentRK, IdentWOO=:IdentWOO, ParentRK=:ParentRK, ParentWOO=:ParentWOO, Sync=:Sync, Status=:Status 
                 WHERE IdentRK=:IdentRK;`
			logger.Debugf("UPDATE:\n%s(%v)", query, c)
			result, err := db.NamedExec(query,
				map[string]interface{}{
					"Name":      c.Name,
					"LongName":  c.LongName,
					"IdentRK":   c.IdentRK,
					"IdentWOO":  c.IdentWOO,
					"ParentRK":  c.ParentRK,
					"ParentWOO": c.ParentWOO,
					"Sync":      c.Sync,
					"Status":    c.Status,
				})
			if err != nil {
				return errors.Wrapf(err, "failed UPDATE to dbsqlite; query:\n%s(%v)", query, c)
			} else {
				l, _ := result.RowsAffected()
				logger.Debugf("Строк %d обновлено успешно", l)
				err := tx.Commit()
				if err != nil {
					return errors.Wrapf(err, "failed in Commit(); failed UPDATE to dbsqlite; query:\n%s(%v)", query, c)
				} else {
					logger.Info("Commit() is done")
					return nil
				}
			}
		} else {
			logger.Debug("Обновление строки не требуется")
			return nil
		}
	case len(сateglistInDb) > 1:
		logger.Debug("Необходимо удалить дублирующие строки с полями IdentRK")
		query = "DELETE FROM Categlist WHERE IdentRK=$1;"
		tx.MustExec(query, c.IdentRK)
		logger.Debugf("DELETE:\n%s(%d)", query, c.IdentRK)
		logger.Debug("Строки удалены успешно")
		logger.Debug("Требуется добавить строку")
		query1 := `INSERT INTO Categlist (Name, LongName, IdentRK, IdentWOO, ParentRK, ParentWOO, Sync, Status) 
			VALUES (:Name, :LongName, :IdentRK, :IdentWOO, :ParentRK, :ParentWOO, :Sync, :Status);`
		logger.Debugf("INSERT:\n%s(%v)", query1, c)
		result, err := tx.NamedExec(query1, c)
		if err != nil {
			return errors.Wrapf(err, "failed INSERT to dbsqlite; query:\n%s(%v)", query, c) // todo add error query1
		} else {
			l, _ := result.RowsAffected()
			logger.Debugf("Строк %d добавлено успешно", l)
			err := tx.Commit()
			if err != nil {
				return errors.Wrapf(err, "failed in Commit(); failed INSERT to dbsqlite; query:\n%s(%v)", query, c) // todo add error query1
			} else {
				logger.Info("Commit() is done")
				return nil
			}
		}
	default:
		return nil
	}
}
