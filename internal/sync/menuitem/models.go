package menuitem

import modelsRK7API "WooWithRkeeper/internal/rk7api/models"

const (
	WOO_PRODUCT_IN_STOCK      = "instock"
	WOO_PRODUCT_OUT_OF_STOCK  = "outofstock"
	WOO_PRODUCT_STATUS_ACTIVE = "publish"
)

const (
	IGNORE           = "IGNORE"           // Обнуляем WOO/RK7 - Игнор-лист
	NOT_ACTIVE       = "NOT_ACTIVE"       // Обнуляем WOO/RK7 - Не активное
	NOT_PRICE        = "NOT_PRICE"        // Обнуляем WOO/RK7 - Не указана цена
	STOP_LIST        = "STOP_LIST"        // Обнуляем WOO/RK7 - В стоп-листе
	NOT_PARENT       = "NOT_PARENT"       // Обнуляем WOO/RK7 - Не найден Parent
	PARENT_SYNC_OFF  = "PARENT_SYNC_OFF"  // Обнуляем WOO/RK7 - Parent с выключенной синхронизацией или не активный, или корневой раздел
	NOT_WOO_ID       = "NOT_WOO_ID"       // Создаем в WOO - Не указан WOO_ID
	NOT_FOUND_IN_WOO = "NOT_FOUND_IN_WOO" // Создаем в WOO - Блюдо не найдено в WOO
	NEED_UPDATE      = "NEED_UPDATE"      // Обновляем WOO - Папка RK7 не совпадает с WOO(свойства Name/LongName/WOO_ID)
	NOT_NEED_UPDATE  = "NOT_NEED_UPDATE"  // Обновление в WOO не требуется
)

type MenuitemSync struct {
	*modelsRK7API.MenuitemItem
	StatusSync string
}
