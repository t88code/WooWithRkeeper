package sync

import (
	"WooWithRkeeper/internal/database"
	"WooWithRkeeper/internal/rk7api"
	"WooWithRkeeper/pkg/logging"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"strings"
)

//func GetVersion()
func GetVersion(rk7api rk7api.RK7API, RefName string) (int, error) {

	logger := logging.GetLogger()
	logger.Println("Start GetVersion")
	defer logger.Println("End GetVersion")

	//получить версию меню из RK7
	RefList, err := rk7api.GetRefList()
	if err != nil {
		return 0, errors.Wrap(err, "Ошибка в Response rk7api.GetRefList")
	}

	var VersionRK7 int
	for _, rk7Reference := range RefList.RK7RefList.RK7Reference {
		if strings.ToLower(rk7Reference.RefName) == strings.ToLower(RefName) {
			VersionRK7 = rk7Reference.DataVersion
			logger.Printf("Версия справочника %s : %d", RefName, VersionRK7)
			break
		}
	}

	if VersionRK7 == 0 {
		return 0, errors.New(fmt.Sprintf("Версия RK справочника %s не определена", RefName))
	}

	return VersionRK7, nil
}

//func VerifyVersion()
func VerifyVersion(rk7api rk7api.RK7API, db *sqlx.DB, RefName string) (bool, error) {

	logger := logging.GetLogger()
	logger.Println("Start VerifyVersion")
	defer logger.Println("End VerifyVersion")

	var err error
	var VersionRK7 int
	if RefName == "Menuitems" || RefName == "Categlist" || RefName == "Prices" {
		VersionRK7, err = GetVersion(rk7api, RefName)
		if err != nil {
			return false, errors.Wrapf(err, "failed GetVersion(rk7api, %s)", RefName)
		}
	} else if RefName == "GetOrderList" {
		resultGetOrderList, err := rk7api.GetOrderList()
		if err != nil {
			return false, errors.Wrapf(err, "failed GetOrderList()")
		}
		VersionRK7 = resultGetOrderList.Lastversion
	}

	//получить версию меню из DB
	type Version database.Version
	var Versions []Version
	query := fmt.Sprintf(`SELECT Version FROM Version WHERE Name='%s'`, RefName)
	err = db.Select(&Versions, query)
	if err != nil {
		return false, errors.Wrap(err, "failed SELECT to dbsqlite")
	}
	var VersionDB int
	if len(Versions) > 0 { // если найдена строка с версией
		VersionDB = Versions[0].Version
		logger.Printf("Версия %s в таблице DB Version: %d", RefName, VersionDB)
	} else { //строки с версией нет, поэтому создаем ее, с нулевой версией
		logger.Printf("Версия %s не определена, т.к. не найдена строка в таблице Version", RefName)
		insert := fmt.Sprintf(`INSERT INTO Version (Name, Version) VALUES ('%s', 0)`, RefName)
		exec := db.MustExec(insert)
		id, err := exec.LastInsertId()
		if err != nil {
			return false, errors.Wrapf(err, "failed INSERT: %s", insert)
		}
		if id == 0 {
			return false, errors.New("INSERT failed, ID = 0 ")
		}
		logger.Println("INSERT OK. ID: ", id)

		affected, err := exec.RowsAffected()
		if err != nil {
			return false, errors.Wrapf(err, "failed INSERT: %s", insert)
		}
		if affected > 1 {
			return false, errors.New(fmt.Sprintf("UPDATE failed, affected = %d ", affected))
		}
		logger.Println("INSERT OK. Rows affected: ", affected)
		logger.Printf("Создана строка для %s в таблице Version", RefName)
	}

	if VersionRK7 == VersionDB {
		logger.Printf("Синхронизация не требуется. Версия %s совпадает между DB и RK7", RefName)
		return true, nil
	} else {
		logger.Printf("Требуется синхронизацию. Версия %s не совпадает между DB и RK7", RefName)
		return false, nil
	}
}

//func UpdateVersionInDB
func UpdateVersionInDB(db *sqlx.DB, RefName string, Version int) error {
	logger := logging.GetLogger()
	logger.Println("Start UpdateVersionInDB")
	defer logger.Println("End UpdateVersionInDB")

	updateQuery := fmt.Sprintf("UPDATE Version SET Version=$1 WHERE Name='$2'")
	exec := db.MustExec(updateQuery, Version, RefName)
	_, err := exec.LastInsertId()
	if err != nil {
		return errors.Wrapf(err, "failed UPDATE: %s", updateQuery)
	}
	logger.Println("UPDATE OK")

	affected, err := exec.RowsAffected()
	if err != nil {
		return errors.Wrapf(err, "failed UPDATE: %s", updateQuery)
	}
	if affected > 1 {
		return errors.New(fmt.Sprintf("UPDATE failed, affected = %d ", affected))
	}
	logger.Println("UPDATE OK. Rows affected: ", affected)
	logger.Printf("Обновлена строка для %s в таблице Version", RefName)

	return nil
}
