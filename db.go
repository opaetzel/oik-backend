package main

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var db *sqlx.DB

var schema = `
CREATE TABLE IF NOT EXISTS units (
	title varchar,
	published boolean,
	rotate_image_id integer,
	user_id integer,
	id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS pages (
	title varchar(255),
	unit_id integer,
	id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS rows (
	left_column_id integer,
	right_column_id integer,
	page_id integer,
	id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS columns (
	image_id integer,
	text_id integer,
	id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS texts (
	markdown text,
	html text,
	id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS rotate_images ( 
	basepath varchar(255),
	num integer,
	caption text,
	credits text,
	id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS images ( 
	path varchar(255),
	caption text,
	credits text,
	id SERIAL PRIMARY KEY
);
`

func initDB(dbname string, user string, pw string) {
	var err error
	db, err = sqlx.Connect("postgres", fmt.Sprintf("dbname=%s user=%s password=%s sslmode=disable", dbname, user, pw))
	if err != nil {
		log.Fatalln(err)
	}

	db.MustExec(schema)
}
