package model

type Version struct {
	ID      int    `db:"ID"`
	Name    string `db:"Name"`
	Version int    `db:"Version"`
}
