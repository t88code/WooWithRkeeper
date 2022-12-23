package imagefile

import "database/sql"

type ImageFile struct {
	ID       int            `db:"ID"`
	Name     sql.NullString `db:"Name"`     //todo перепроверить переменные
	ModTime  sql.NullString `db:"ModTime"`  //todo перепроверить переменные
	IdentWOO sql.NullInt32  `db:"IdentWOO"` //todo проверить это 0 или null, когда не указан
}
