package sync

import (
	_ "github.com/mattn/go-sqlite3"
)

const (
	DB_NAME_SQLITE                   = "db.db"
	ERROR_PRODUCT_NOT_FOUND          = "API BX24: error_description: Product is not found; error: "
	ERROR_PRODUCTSECTION_NOT_FOUND   = "API BX24: error_description: Раздел не найден.; error: "
	ERROR_PRODUCCATEGORIES_NOT_FOUND = "code:woocommerce_rest_term_invalid; message:Ресурса не существует.; status:404; display:; details:;"
	ERROR_PRODUCCATEGORIES_IS_EXIST  = "code:term_exists; message:Элемент с указанным именем уже существует у родительского элемента.; status:400; display:; details:;"
)

// TODO сделать ручник! на обновление меню
