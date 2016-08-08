package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var db *sqlx.DB

var schema = `
CREATE TABLE IF NOT EXISTS units (
	unit_title varchar,
	published boolean,
	rotate_image_id integer,
	user_id integer,
	unit_id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS pages (
	page_title varchar(255),
	unit_id integer,
	page_id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS rows (
	left_column_id integer,
	right_column_id integer,
	page_id integer,
	row_id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS columns (
	image_id integer,
	text_id integer,
	column_id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS texts (
	markdown text,
	html text,
	text_id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS rotate_images ( 
	basepath varchar(255),
	num integer,
	caption text,
	credits text,
	rotate_image_id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS images ( 
	path varchar(255),
	caption text,
	credits text,
	image_id SERIAL PRIMARY KEY
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

func GetAllUnits() ([]Unit, error) {
	rows, err := db.Query("SELECT units.*, json_agg(pages.page_id) AS pages_arr FROM units LEFT OUTER JOIN pages ON units.unit_id = pages.unit_id GROUP BY units.unit_id;")
	if err != nil {
		return nil, err
	}
	units := make([]Unit, 0)
	for rows.Next() {
		var unit_title string
		var published bool
		var rotate_image_id int
		var user_id int
		var unit_id int
		var pages_arr string

		err = rows.Scan(&unit_title, &published, &rotate_image_id, &user_id, &unit_id, &pages_arr)
		if err != nil {
			return nil, err
		}
		var pages []int
		err = json.Unmarshal([]byte(pages_arr), &pages)
		if err != nil {
			return nil, err
		}
		units = append(units, Unit{unit_title, nil, pages, published, user_id, unit_id})
	}
	return units, nil
}
