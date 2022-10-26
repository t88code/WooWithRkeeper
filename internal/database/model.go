package database

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
`

const DB_SCHEMA_OLD = `CREATE TABLE Orders (
	ID integer PRIMARY KEY AUTOINCREMENT
);

CREATE TABLE MenuRK7 (
	ID integer PRIMARY KEY AUTOINCREMENT,
	Ident integer,
	Code integer,
	Name text,
	RegularPrice integer,
	MainParentIdent integer,
	Status integer,
	ID_BX24 integer
);

CREATE TABLE MenuBX24 (
	ID integer PRIMARY KEY AUTOINCREMENT,
	Name text,
	RegularPrice integer
);

CREATE TABLE CateglistRK7 (
	ID integer PRIMARY KEY AUTOINCREMENT,
	Name text,
	Ident integer,
	Status integer,
	Sync integer
);

CREATE TABLE Version (
	ID integer PRIMARY KEY AUTOINCREMENT,
	Name text,
	Version integer
);
`
