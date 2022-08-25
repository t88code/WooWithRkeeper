package sync

import (
	"WooWithRkeeper/internal/database"
	"WooWithRkeeper/internal/rk7api"
	modelsRK7API "WooWithRkeeper/internal/rk7api/models"
	"WooWithRkeeper/pkg/logging"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"os"
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
	if RefName == "Menuitems" || RefName == "Categlist" {
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

	updateQuery := fmt.Sprintf(`UPDATE Version SET Version=%d WHERE Name='%s'`, Version, RefName)
	exec := db.MustExec(updateQuery)
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

func GenerateMenuInRK7Map(rk7api rk7api.RK7API) (map[int]modelsRK7API.MenuitemItem, error) {

	logger := logging.GetLogger()
	logger.Println("Start GenerateMenuInRK7Map")
	defer logger.Println("End GenerateMenuInRK7Map")

	//получить список всех блюд из RK
	logger.Println("Получить список всех блюд из RK7")
	Rk7QueryResultGetRefDataMenuitems, err := rk7api.GetRefData("Menuitems",
		modelsRK7API.OnlyActive("0"), //неактивные будем менять статус на N ?или может удалять? в bitrix24
		modelsRK7API.IgnoreEnums("1"),
		modelsRK7API.WithChildItems("3"),
		modelsRK7API.WithMacroProp("1"),
		modelsRK7API.PropMask("items.(Code,Name,Ident,ItemIdent,GUIDString,MainParentIdent,ExtCode,PRICETYPES^3,CategPath,Status,genIDBX24,genSectionIDBX24)"))
	if err != nil {
		return nil, errors.Wrap(err, "Ошибка при выполнении rk7api.GetRefData")
	}

	// PRICETYPES-3="9223372036854775807"

	MenuInRK7 := (Rk7QueryResultGetRefDataMenuitems).(*modelsRK7API.RK7QueryResultGetRefDataMenuitems)

	logger.Printf("Длина списка MenuInRK7 = %d\n", len(MenuInRK7.RK7Reference.Items.Item))

	MenuInRK7Map := make(map[int]modelsRK7API.MenuitemItem)
	for i, item := range MenuInRK7.RK7Reference.Items.Item {
		MenuInRK7Map[item.ItemIdent] = MenuInRK7.RK7Reference.Items.Item[i]
	}

	logger.Printf("Длина списка MenuInRK7Map = %d\n", len(MenuInRK7Map))
	logger.Printf("MenuInRK7Map успешно создан")

	return MenuInRK7Map, nil
}

//func UpdateOrdersInDB
func UpdateOrdersInDB(db *sqlx.DB, RK_VisitID int, RK_GUID string, RK_Deleted int, RK_Version int, BX24_DealID int, BX24_Title string, BX24_DATE_MODIFY string, Sum string, FC_Chmode int) error {
	logger := logging.GetLogger()
	logger.Println("Start UpdateOrdersInDB")
	defer logger.Println("End UpdateOrdersInDB")

	//проверить, что orders не существует
	//получить orders из DB
	type Order database.Order
	var Orders []Order
	query := fmt.Sprintf(`SELECT ID, RK_VisitID, RK_GUID, RK_Deleted, RK_Version, BX24_DealID, BX24_Title, BX24_DATE_MODIFY, Sum, FC_Chmode, Sync FROM Version WHERE RK_VisitID='%d'`, RK_VisitID)
	logger.Debugf("Query: %s", query)
	err := db.Select(&Orders, query)
	if err != nil {
		return errors.Wrapf(err, "failed SELECT to dbsqlite, query: %s", query)
	}

	if len(Orders) == 0 {
		//если не сущесвтует, то создать
		logger.Info("Visit не найден в таблице Orders")
		insert := fmt.Sprintf(`INSERT INTO Orders (RK_VisitID, RK_GUID, RK_Deleted, RK_Version, BX24_DealID, BX24_Title, BX24_DATE_MODIFY, Sum, FC_Chmode, Sync) VALUES ('%d', '%s', '%d', '%d', '%d', '%s', '%s', '%s', '%d', '%d')`,
			RK_VisitID, RK_GUID, RK_Deleted, RK_Version, BX24_DealID, BX24_Title, BX24_DATE_MODIFY, Sum, FC_Chmode, 1)
		exec := db.MustExec(insert)
		id, err := exec.LastInsertId()
		if err != nil {
			return errors.Wrapf(err, "failed INSERT: %s", insert)
		}
		if id == 0 {
			return errors.New("INSERT failed, ID = 0 ")
		}
		affected, err := exec.RowsAffected()
		if err != nil {
			return errors.Wrapf(err, "failed INSERT: %s", insert)
		}
		if affected > 1 {
			return errors.New(fmt.Sprintf("UPDATE failed, affected = %d ", affected))
		}
		logger.Info("INSERT OK. Rows affected: ", affected)
		logger.Infof("Создана строка для %d в таблице Orders", RK_VisitID)

	} else if len(Orders) == 1 {
		//если существует, то обновить

	} else {
		return errors.New(fmt.Sprintf("найдено несколько VisitID=%d в таблице Orders, что недопустимо", RK_VisitID))
	}

	return nil
}

// проверка, существует ли файл или нет
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func CreateDB() error {

	logger := logging.GetLogger()
	logger.Print("CreateDB:>Start")
	defer logger.Print("CreateDB:>End")

	logger.Print("CreateDB:>Creating ", DB_NAME_SQLITE)

	db, err := sqlx.Open("sqlite3", DB_NAME_SQLITE)
	if err != nil {
		logger.Fatal(err)
		return err
	}
	defer db.Close()

	logger.Print(DB_NAME_SQLITE, " created")

	db.MustExec(database.DB_SCHEMA)
	return nil
}
