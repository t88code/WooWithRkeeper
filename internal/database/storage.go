package database

const DB_NAME = "db.db"

const DB_SCHEMA = `CREATE TABLE Version (
	ID integer PRIMARY KEY AUTOINCREMENT,
	Name text,
	Version integer
);

CREATE TABLE Image (
	ID integer PRIMARY KEY AUTOINCREMENT,
	IdentRK integer,
	ModTime text,
	Name text,
	Pos integer, 
	IdentWOO integer,
	Status text
);
`
