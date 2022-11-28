package database

import "database/sql"

type Version struct {
	ID      int    `db:"ID"`
	Name    string `db:"Name"`
	Version int    `db:"Version"`
}

type Order struct {
	ID               int    `db:"ID"`
	RK_VISITID       int    `db:"RK_VisitID"`
	RK_GUID          string `db:"RK_GUID"`
	RK_DELETED       int    `db:"RK_Deleted"`
	RK_VERSION       int    `db:"RK_Version"`
	BX24_DEAL_ID     int    `db:"BX24_DealID"`
	BX24_TITLE       string `db:"BX24_Title"`
	BX24_DATE_MODIFY string `db:"BX24_DATE_MODIFY"`
	SUM              string `db:"Sum"`
	FC_CHMODE        int    `db:"FC_Chmode"`
	SYNC             int    `db:"Sync"`
}

type Menuitem struct {
	ID               int            `db:"ID"`
	Ident            int            `db:"IdentRK"`
	IMAGE_MOD_TIME_1 sql.NullString `db:"ImageModTime1"`
	IMAGE_NAME_1     sql.NullString `db:"ImageName1"`
	IMAGE_MOD_TIME_2 sql.NullString `db:"ImageModTime2"`
	IMAGE_NAME_2     sql.NullString `db:"ImageName2"`
}

const DATABASE_SELECT_MENUITEM = `SELECT ID, IdentRK, ImageModTime1, ImageName1, ImageModTime2, ImageName2 FROM Menuitem WHERE IdentRK=%d`

const DB_SCHEMA = `CREATE TABLE Orders (
	ID integer PRIMARY KEY AUTOINCREMENT,
	RK_VisitID integer,
	RK_GUID text,
	RK_Deleted integer,
	RK_Version integer,
	BX24_DealID integer,
	BX24_Title text,
	BX24_DATE_MODIFY text,
	Sum text,
	FC_Chmode integer,
	Sync integer
);

CREATE TABLE Version (
	ID integer PRIMARY KEY AUTOINCREMENT,
	Name text,
	Version integer
);

CREATE TABLE Menuitem (
	ID integer PRIMARY KEY AUTOINCREMENT,
	IdentRK integer,
	ImageModTime1 text,
	ImageName1 text,
	ImageModTime2 text,
	ImageName2 text
);
`
