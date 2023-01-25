package database

const DB_SCHEMA = `CREATE TABLE Version (
	ID integer PRIMARY KEY AUTOINCREMENT,
	Name text,
	Version integer
);

CREATE TABLE ImageFile (
	ID integer PRIMARY KEY AUTOINCREMENT,
	Name text,
	ModTime text,
	IdentWOO integer
);

CREATE TABLE Categlist (
	ID integer PRIMARY KEY AUTOINCREMENT,
	Name text,
	LongName text,
	IdentRK integer,
	IdentWOO integer,
	ParentRK integer,
	ParentWOO integer,
	Sync integer,
	Status text
);
`
