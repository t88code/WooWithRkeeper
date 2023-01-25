package categlist

import modelsRK7API "WooWithRkeeper/internal/rk7api/models"

const (
	IGNORE           = "IGNORE"           // Обнулить WOO/RK7 - Игнор-лист
	SYNC_OFF         = "SYNC_OFF"         // Обнуляем в WOO/RK7 - Синхронизация отключена
	NOT_ACTIVE       = "NOT_ACTIVE"       // Обнуляем в WOO/RK7 - Папка не активная
	NOT_WOO_ID       = "NOT_WOO_ID"       // Создаем в WOO - Не указан WOO_ID
	NOT_FOUND_IN_WOO = "NOT_FOUND_IN_WOO" // Создаем в WOO - Папка не найдена в WOO
	NEED_UPDATE      = "NEED_UPDATE"      // Обновляем WOO - Папка RK7 не совпадает с WOO(свойства Name/LongName)
	NOT_NEED_UPDATE  = "NOT_NEED_UPDATE"  // Обновление в WOO не требуется - Папка RK7 совпадает с WOO(свойства Name/LongName)
)

type CateglistSync struct {
	*modelsRK7API.Categlist
	StatusSync string
}
