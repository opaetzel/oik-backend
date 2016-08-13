package main

import (
	"database/sql"
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
	unit_id integer,
	image_id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS users (
	username varchar(255),
	salt varchar(255),
	pwhash varchar(255),
	active boolean,
	user_id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS groups (
	group_name varchar(255),
	group_id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS user_groups (
	user_id integer,
	group_id integer
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

func parseUnits(rows *sql.Rows) ([]Unit, error) {
	units := make([]Unit, 0)
	for rows.Next() {
		var unit_title string
		var published bool
		var rotate_image_id int
		var user_id int
		var unit_id int
		var pages_arr string

		err := rows.Scan(&unit_title, &published, &rotate_image_id, &user_id, &unit_id, &pages_arr)
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

func GetAllUnits() ([]Unit, error) {
	rows, err := db.Query("SELECT units.*, json_agg(pages.page_id) AS pages_arr FROM units LEFT OUTER JOIN pages ON units.unit_id = pages.unit_id GROUP BY units.unit_id;")
	if err != nil {
		return nil, err
	}
	return parseUnits(rows)
}

func GetUserUnits(userId int) ([]Unit, error) {
	rows, err := db.Query("SELECT units.*, json_agg(pages.page_id) AS pages_arr FROM units LEFT OUTER JOIN pages ON units.unit_id = pages.unit_id WHERE units.user_id=$1 GROUP BY units.unit_id;", userId)
	if err != nil {
		return nil, err
	}
	return parseUnits(rows)
}

func GetPublishedUnits() ([]Unit, error) {
	rows, err := db.Query("SELECT units.*, json_agg(pages.page_id) AS pages_arr FROM units LEFT OUTER JOIN pages ON units.unit_id = pages.unit_id WHERE units.published=true GROUP BY units.unit_id;")
	if err != nil {
		return nil, err
	}
	return parseUnits(rows)
}

func GetPageOwner(pageId int) (int, error) {
	query := `
		SELECT units.user_id FROM units
		JOIN pages ON pages.unit_id = units.unit_id
		WHERE pages.page_id = $1;
		`
	row := db.QueryRow(query, pageId)
	var userId int
	err := row.Scan(&userId)
	if err != nil {
		return -1, err
	}
	return userId, nil
}

func UpdatePage(page Page) error {
	stmt, err := db.Prepare("UPDATE pages SET (page_title) VALUES ($1);")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(page.Title)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err = tx.Prepare("UPDATE rows SET (left_markdown, left_html, right_markdown, right_html) VALUES ($1, $2, $3, $4);")
	for _, row := range page.Rows {
		_, err := stmt.Exec(row.LeftMarkdown, row.LeftHtml, row.RightMarkdown, row.RightHtml)
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

func parsePage(row *sql.Row) (Page, error) {
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

func GetPublicPageById(id int) (Page, error) {
	query := `
		SELECT pages.page_title, pages.page_id, pages.unit_id, json_agg(rows.*) AS rows FROM pages 
		LEFT JOIN rows ON rows.page_id = pages.page_id
		RIGHT JOIN units ON units.unit_id = pages.unit_id
		WHERE pages.page_id=$1 AND units.published = true
		GROUP BY pages.page_id;
		`
	row := db.QueryRow(query, id)
	return parsePage(row)
}

func GetPageById(id int) (Page, error) {
	query := `
		SELECT pages.page_title, pages.page_id, pages.unit_id, json_agg(rows.*) AS rows FROM pages 
		LEFT JOIN rows ON rows.page_id = pages.page_id
		RIGHT JOIN units ON units.unit_id = pages.unit_id
		WHERE pages.page_id=$1
		GROUP BY pages.page_id;
		`
	row := db.QueryRow(query, id)
	return parsePage(row)
}

func GetUserPageById(pageId, userId int) (Page, error) {
	query := `
		SELECT pages.page_title, pages.page_id, pages.unit_id, json_agg(rows.*) AS rows FROM pages 
		LEFT JOIN rows ON rows.page_id = pages.page_id
		RIGHT JOIN units ON units.unit_id = pages.unit_id
		WHERE pages.page_id=$1 AND units.user_id = $2
		GROUP BY pages.page_id;
		`
	row := db.QueryRow(query, pageId, userId)
	return parsePage(row)
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

func UpdateUnitAdmin(unit Unit) error {
	stmt, err := db.Prepare("UPDATE units SET (unit_title, published, rotate_image_id) VALUES ($1, $2, $3);")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(unit.Title, unit.Published, unit.UnitImageID)
	if err != nil {
		return err
	}
	return nil
}

func UpdateUnitUser(unit Unit) error {
	stmt, err := db.Prepare("UPDATE units SET (unit_title, rotate_image_id) VALUES ($1, $2);")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(unit.Title, unit.UnitImageID)
	if err != nil {
		return err
	}
	return nil
}

func GetImageOwner(imageId int) (int, error) {
	query := `
		SELECT units.user_id FROM units
		JOIN images ON images.unit_id = units.unit_id
		WHERE images.image_id = $1;
		`
	row := db.QueryRow(query, imageId)
	var userId int
	err := row.Scan(&userId)
	if err != nil {
		return -1, err
	}
	return userId, nil
}

func InsertImage(image Image) (int, error) {
	query := "INSERT INTO images (caption, credits, unit_id) VALUES ($1, $2, $3) RETURNING image_id;"
	var imageId int
	err := db.QueryRow(query, image.Caption, image.Credits, image.UnitId).Scan(&imageId)
	if err != nil {
		return -1, err
	}
	return int(imageId), nil
}

func InsertPage(page Page) error {
	query := "INSERT INTO pages (page_title, unit_id) VALUES ($1, $2) RETURNING page_id;"
	var pageId int
	err := db.QueryRow(query, page.Title, page.UnitID).Scan(&pageId)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO rows (left_markdown, left_html, right_markdown, right_html, page_id) VALUES ($1, $2, $3, $4, $5);")
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

func GetUser(username string) (User, error) {
	row := db.QueryRow("SELECT users.*, json_agg(groups.group_name) FROM users LEFT JOIN user_groups ON users.user_id=user_groups.user_id LEFT JOIN groups ON user_groups.group_id=groups.group_id WHERE username=$1 GROUP BY users.user_id", username)
	var dbUsername, salt, pwhash, jsonGroups string
	var active bool
	var id int
	err := row.Scan(&dbUsername, &salt, &pwhash, &active, &id, &jsonGroups)
	if err != nil {
		return User{}, err
	}
	var groups []string
	if err := json.Unmarshal([]byte(jsonGroups), &groups); err != nil {
		return User{}, err
	}
	u := User{dbUsername, groups, salt, pwhash, active, id}
	return u, nil
}

func InsertUser(user User) error {
	stmt, err := db.Prepare("INSERT INTO users (username, salt, pwhash, active) VALUES ($1, $2, $3, $4);")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(user.Username, user.Salt, user.PWHash, user.Active)
	if err != nil {
		return err
	}
	return nil
}
