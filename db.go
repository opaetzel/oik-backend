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
	left_markdown text,
	left_html text,
	right_markdown text,
	right_html text,
	page_id integer,
	row_id SERIAL PRIMARY KEY
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
		units = append(units, Unit{unit_title, rotate_image_id, pages, published, user_id, unit_id})
	}
	return units, nil
}

func GetPageById(id int) (Page, error) {
	query := `
		SELECT pages.page_title, pages.page_id, pages.unit_id, json_agg(rows.*) AS rows FROM pages 
		LEFT JOIN rows ON rows.page_id = pages.page_id
		WHERE pages.page_id=$1
		GROUP BY pages.page_id;
		`
	row := db.QueryRow(query, id)
	var pageTitle, jsonRows string
	var pageId, unitId int
	if err := row.Scan(&pageTitle, &pageId, &unitId, &jsonRows); err != nil {
		return Page{}, err
	}
	var rows []Row
	if err := json.Unmarshal([]byte(jsonRows), &rows); err != nil {
		return Page{}, err
	}
	return Page{pageTitle, rows, unitId, pageId}, nil
}

func InsertUnit(unit Unit) error {
	stmt, err := db.Prepare("INSERT INTO units (unit_title, published, rotate_image_id, user_id) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(unit.Title, unit.Published, unit.UnitImageID, unit.UserId)
	if err != nil {
		return err
	}
	return nil
}

func InsertPage(page Page) error {
	stmt, err := db.Prepare("INSERT INTO pages (page_title, unit_id) VALUES ($1, $2);")
	if err != nil {
		return err
	}
	res, err := stmt.Exec(page.Title, page.UnitID)
	if err != nil {
		return err
	}
	pageId, err := res.LastInsertId()
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err = tx.Prepare("INSERT INTO rows (left_markdown, left_html, right_markdown, right_html, page_id) VALUES ($1, $2, $3, $4, $5);")
	for _, row := range page.Rows {
		_, err := stmt.Exec(row.LeftMarkdown, row.LeftHtml, row.RightMarkdown, row.RightHtml, pageId)
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
