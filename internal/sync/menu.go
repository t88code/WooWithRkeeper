package sync

import (
	_ "github.com/mattn/go-sqlite3"
)

const (
	DB_NAME_SQLITE                 = "db.db"
	ERROR_PRODUCT_NOT_FOUND        = "API BX24: error_description: Product is not found; error: " //TODO
	ERROR_PRODUCTSECTION_NOT_FOUND = "API BX24: error_description: Раздел не найден.; error: "    //TODO
)

//func SyncMenuService()

// TODO сделать ручник! на обновление меню
