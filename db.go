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

var emptyArr = "[null]"

var schema = `
CREATE TABLE IF NOT EXISTS units (
	unit_title varchar,
	published boolean,
	rotate_image_id integer,
	user_id integer,
	color_scheme integer,
	unit_id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS pages (
	page_title varchar(255),
	unit_id integer,
	page_type varchar(255),
	page_id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS rows (
	left_markdown text,
	right_markdown text,
	left_has_image boolean,
	right_has_image boolean,
	leftImage integer,
	rightImage integer,
	left_is_argument boolean,
	right_is_argument boolean,
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
	user_id SERIAL PRIMARY KEY,
	mailhash varchar(255)
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
		var color_scheme int
		var unit_id int
		var pages_arr, images_arr string

		err := rows.Scan(&unit_title, &published, &rotate_image_id, &user_id, &color_scheme, &unit_id, &pages_arr, &images_arr)
		if err != nil {
			return nil, err
		}
		var pages []int
		if pages_arr == emptyArr {
			pages = make([]int, 0)
		} else {
			err = json.Unmarshal([]byte(pages_arr), &pages)
			if err != nil {
				return nil, err
			}
		}
		var images []int
		if images_arr == emptyArr {
			images = make([]int, 0)
		} else {
			err = json.Unmarshal([]byte(images_arr), &images)
			if err != nil {
				return nil, err
			}
		}
		units = append(units, Unit{unit_title, rotate_image_id, pages, published, color_scheme, user_id, images, unit_id})
	}
	return units, nil
}

func parseUnit(row *sql.Row) (Unit, error) {
	var unit_title string
	var published bool
	var rotate_image_id int
	var user_id int
	var color_scheme int
	var unit_id int
	var pages_arr, images_arr string

	err := row.Scan(&unit_title, &published, &rotate_image_id, &user_id, &color_scheme, &unit_id, &pages_arr, &images_arr)
	if err != nil {
		return Unit{}, err
	}
	var pages []int
	if pages_arr == emptyArr {
		pages = make([]int, 0)
	} else {
		err = json.Unmarshal([]byte(pages_arr), &pages)
		if err != nil {
			return Unit{}, err
		}
	}
	var images []int
	if images_arr == emptyArr {
		images = make([]int, 0)
	} else {
		err = json.Unmarshal([]byte(images_arr), &images)
		if err != nil {
			return Unit{}, err
		}
	}
	return Unit{unit_title, rotate_image_id, pages, published, color_scheme, user_id, images, unit_id}, nil
}

func GetUnit(unitId int) (Unit, error) {
	row := db.QueryRow("SELECT units.*, json_agg(DISTINCT pages.page_id) AS pages_arr, json_agg(DISTINCT images.image_id) AS images_arr FROM units LEFT OUTER JOIN pages ON units.unit_id = pages.unit_id LEFT OUTER JOIN images ON units.unit_id=images.unit_id WHERE units.unit_id=$1 GROUP BY units.unit_id;", unitId)
	return parseUnit(row)
}

func GetAllUnits() ([]Unit, error) {
	rows, err := db.Query("SELECT units.*, json_agg(DISTINCT pages.page_id) AS pages_arr, json_agg(DISTINCT images.image_id) AS images_arr FROM units LEFT OUTER JOIN pages ON units.unit_id = pages.unit_id LEFT OUTER JOIN images ON units.unit_id=images.unit_id GROUP BY units.unit_id;")
	if err != nil {
		return nil, err
	}
	return parseUnits(rows)
}

/*
func GetUserUnits(userId int) ([]Unit, error) {
	rows, err := db.Query("SELECT units.*, json_agg(pages.page_id) AS pages_arr FROM units LEFT OUTER JOIN pages ON units.unit_id = pages.unit_id WHERE units.user_id=$1 GROUP BY units.unit_id;", userId)
	if err != nil {
		return nil, err
	}
	return parseUnits(rows)
}
*/
func GetUnPublishedUnits() ([]Unit, error) {
	rows, err := db.Query("SELECT units.*, json_agg(DISTINCT pages.page_id) AS pages_arr, json_agg(DISTINCT images.image_id) FROM units LEFT OUTER JOIN pages ON units.unit_id = pages.unit_id LEFT OUTER JOIN images ON units.unit_id=images.unit_id WHERE units.published=false GROUP BY units.unit_id;")
	if err != nil {
		return nil, err
	}
	return parseUnits(rows)
}

func GetPublishedUnits() ([]Unit, error) {
	rows, err := db.Query("SELECT units.*, json_agg(DISTINCT pages.page_id) AS pages_arr, json_agg(DISTINCT images.image_id) FROM units LEFT OUTER JOIN pages ON units.unit_id = pages.unit_id LEFT OUTER JOIN images ON units.unit_id=images.unit_id WHERE units.published=true GROUP BY units.unit_id;")
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
	stmt, err := db.Prepare("UPDATE pages SET page_title=$1, page_type=$2 WHERE page_id=$3;")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(page.Title, page.PageType, page.ID)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err = tx.Prepare("UPDATE rows SET left_markdown=$1, right_markdown=$2, left_has_image=$3, right_has_image=$4 leftimage=$5, rightimage=$6, left_is_argument=$7, right_is_argument=$8, WHERE row_id=$9;")
	for _, row := range page.Rows {
		_, err := stmt.Exec(row.LeftMarkdown, row.RightMarkdown, row.LeftHasImage, row.RightHasImage, row.LeftImage, row.RightImage, row.LeftIsArgument, row.RightIsArgument, row.ID)
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
	var published bool
	var pageTitle, page_type, jsonRows string
	var pageId, unitId, userId int
	if err := row.Scan(&published, &userId, &pageTitle, &pageId, &unitId, &page_type, &jsonRows); err != nil {
		return Page{}, err
	}
	var rows []Row
	if jsonRows == emptyArr {
		rows = make([]Row, 0)
	} else {
		if err := json.Unmarshal([]byte(jsonRows), &rows); err != nil {
			return Page{}, err
		}
	}
	return Page{pageTitle, rows, unitId, page_type, pageId, userId, published}, nil
}

/*
func GetPublicPageById(id int) (Page, error) {
	query := `
		SELECT pages.page_title, pages.page_id, pages.unit_id, pages.page_type, json_agg(rows.*) AS rows FROM pages
		LEFT JOIN rows ON rows.page_id = pages.page_id
		RIGHT JOIN units ON units.unit_id = pages.unit_id
		WHERE pages.page_id=$1 AND units.published = true
		GROUP BY pages.page_id;
		`
	row := db.QueryRow(query, id)
	return parsePage(row)
}
*/
func GetPageById(id int) (Page, error) {
	query := `
		SELECT units.published, units.user_id, pages.page_title, pages.page_id, pages.unit_id, pages.page_type, json_agg(rows.*) AS rows FROM pages 
		LEFT JOIN rows ON rows.page_id = pages.page_id
		RIGHT JOIN units ON units.unit_id = pages.unit_id
		WHERE pages.page_id=$1
		GROUP BY pages.page_id, units.unit_id;
		`
	row := db.QueryRow(query, id)
	return parsePage(row)
}

/*
func GetUserPageById(pageId, userId int) (Page, error) {
	query := `
		SELECT pages.page_title, pages.page_id, pages.unit_id, pages.page_type, json_agg(rows.*) AS rows FROM pages
		LEFT JOIN rows ON rows.page_id = pages.page_id
		RIGHT JOIN units ON units.unit_id = pages.unit_id
		WHERE pages.page_id=$1 AND units.user_id = $2
		GROUP BY pages.page_id;
		`
	row := db.QueryRow(query, pageId, userId)
	return parsePage(row)
}
*/
func InsertUnit(unit Unit) (int, error) {
	log.Println(unit.Title)
	row := db.QueryRow("INSERT INTO units (unit_title, published, rotate_image_id, user_id, color_scheme) VALUES ($1, $2, $3, $4, $5) RETURNING units.unit_id", unit.Title, unit.Published, unit.UnitImageID, unit.UserId, unit.ColorScheme)
	var id int
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func UpdateUnitAdmin(unit Unit) error {
	stmt, err := db.Prepare("UPDATE units SET unit_title=$1, published=$2, rotate_image_id=$3, color_scheme=$4;")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(unit.Title, unit.Published, unit.UnitImageID, unit.ColorScheme)
	if err != nil {
		return err
	}
	return nil
}

func UpdateUnitUser(unit Unit) error {
	stmt, err := db.Prepare("UPDATE units SET unit_title=$1, rotate_image_id=$2, color_scheme=$3;")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(unit.Title, unit.UnitImageID, unit.ColorScheme)
	if err != nil {
		return err
	}
	return nil
}

func UpdateImagePath(imageId int, imagePath string) error {
	stmt, err := db.Prepare("UPDATE images SET path=$1 WHERE image_id=$2")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(imagePath, imageId)
	if err != nil {
		return err
	}
	return nil
}

func GetRotateImagePublishedAndPath(imageId int) (bool, string, error) {
	query := `
		SELECT units.published, rotate_images.basepath FROM units
		JOIN rotate_images ON rotate_images.rotate_image_id = units.rotate_image_id
		WHERE rotate_images.rotate_image_id = $1;
		`
	row := db.QueryRow(query, imageId)
	var published bool
	var path string
	err := row.Scan(&published, &path)
	if err != nil {
		return false, "", err
	}
	return published, path, nil
}

func GetImageById(imageId int) (Image, error) {
	query := `
		SELECT units.published, units.user_id, images.path, images.caption, images.credits, images.unit_id FROM units
		JOIN images ON images.unit_id = units.unit_id
		WHERE images.image_id = $1;
		`
	row := db.QueryRow(query, imageId)
	var published bool
	var userId, unitId int
	var path, caption, credits string
	var nullPath sql.NullString
	err := row.Scan(&published, &userId, &nullPath, &caption, &credits, &unitId)
	if err != nil {
		return Image{}, err
	}
	if nullPath.Valid {
		path = nullPath.String
	}
	return Image{path, caption, credits, unitId, userId, imageId, published}, nil
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

func InsertRotateImage(image RotateImage) (int, error) {
	query := "INSERT INTO rotate_images (caption, credits, num) VALUES ($1, $2, $3) RETURNING image_id;"
	var imageId int
	err := db.QueryRow(query, image.Caption, image.Credits, image.Num).Scan(&imageId)
	if err != nil {
		return -1, err
	}
	return int(imageId), nil
}

func InsertPage(page Page) (Page, error) {
	query := "INSERT INTO pages (page_title, page_type, unit_id) VALUES ($1, $2, $3) RETURNING page_id;"
	var pageId int
	err := db.QueryRow(query, page.Title, page.PageType, page.UnitID).Scan(&pageId)
	if err != nil {
		return Page{}, err
	}
	page.ID = pageId
	tx, err := db.Begin()
	if err != nil {
		return Page{}, err
	}
	stmt, err := tx.Prepare("INSERT INTO rows (left_markdown, right_markdown, left_has_image, right_has_image, leftimage, rightimage, left_is_argument, right_is_argument, page_id) VALUES ($1, $2, $3, $4, $5 ,$6, $7, $8, $9) RETURNING row_id;")
	if err != nil {
		return Page{}, err
	}
	for idx, row := range page.Rows {
		var rowId int
		err := stmt.QueryRow(row.LeftMarkdown, row.RightMarkdown, row.LeftHasImage, row.RightHasImage, row.LeftImage, row.RightImage, row.LeftIsArgument, row.RightIsArgument, pageId).Scan(&rowId)
		if err != nil {
			return Page{}, err
		}
		log.Println(rowId)
		page.Rows[idx].ID = rowId
	}
	log.Println(page.Rows)
	err = tx.Commit()
	if err != nil {
		return Page{}, err
	}
	return page, nil
}

func GetUserById(userId int) (User, error) {
	query := `SELECT users.username, json_agg(DISTINCT units.unit_id) AS units, json_agg(DISTINCT groups.group_name) FROM users 
		LEFT JOIN units ON units.user_id=users.user_id 
		LEFT JOIN user_groups ON users.user_id=user_groups.user_id 
		LEFT JOIN groups ON user_groups.group_id = groups.group_id
		WHERE users.user_id=$1 GROUP BY users.user_id`
	row := db.QueryRow(query, userId)
	var dbUsername, jsonUnits, jsonGroups string
	err := row.Scan(&dbUsername, &jsonUnits, &jsonGroups)
	if err != nil {
		return User{}, err
	}
	var groups []string
	if err := json.Unmarshal([]byte(jsonGroups), &groups); err != nil {
		return User{}, err
	}
	var units []int
	if jsonUnits == emptyArr {
		units = make([]int, 0)
	} else {
		if err := json.Unmarshal([]byte(jsonUnits), &units); err != nil {
			return User{}, err
		}
	}
	u := User{Username: dbUsername, Units: units, Groups: groups, ID: userId}
	return u, nil
}

func GetUserByName(username string) (User, error) {
	row := db.QueryRow("SELECT users.*, json_agg(groups.group_name) FROM users LEFT JOIN user_groups ON users.user_id=user_groups.user_id LEFT JOIN groups ON user_groups.group_id=groups.group_id WHERE username=$1 GROUP BY users.user_id", username)
	var dbUsername, salt, pwhash, jsonGroups, mailHash string
	var active bool
	var id int
	err := row.Scan(&dbUsername, &salt, &pwhash, &active, &id, &mailHash, &jsonGroups)
	if err != nil {
		return User{}, err
	}
	var groups []string
	if err := json.Unmarshal([]byte(jsonGroups), &groups); err != nil {
		return User{}, err
	}
	u := User{dbUsername, groups, nil, id, salt, pwhash, active, mailHash}
	return u, nil
}

func InsertUser(user User) (int, error) {
	query := "INSERT INTO users (username, salt, pwhash, active, mailhash) VALUES ($1, $2, $3, $4, $5) RETURNING user_id;"
	var userId int
	err := db.QueryRow(query, user.Username, user.salt, user.pwHash, user.Active, user.mailHash).Scan(&userId)
	if err != nil {
		return -1, err
	}
	return userId, nil
}

func GetAllUsers() ([]User, error) {
	query := `SELECT users.username, users.active, users.user_id, json_agg(DISTINCT units.unit_id) AS units, json_agg(DISTINCT groups.group_name) FROM users 
		LEFT JOIN units ON units.user_id=users.user_id 
		LEFT JOIN user_groups ON users.user_id=user_groups.user_id 
		LEFT JOIN groups ON user_groups.group_id = groups.group_id
		GROUP BY users.user_id`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	users := make([]User, 0)
	for rows.Next() {
		var dbUsername, jsonUnits, jsonGroups string
		var active bool
		var userId int
		err := rows.Scan(&dbUsername, &active, &userId, &jsonUnits, &jsonGroups)
		if err != nil {
			return nil, err
		}
		var groups []string
		if err := json.Unmarshal([]byte(jsonGroups), &groups); err != nil {
			return nil, err
		}
		var units []int
		if jsonUnits == emptyArr {
			units = make([]int, 0)
		} else {
			if err := json.Unmarshal([]byte(jsonUnits), &units); err != nil {
				return nil, err
			}
		}
		users = append(users, User{Username: dbUsername, Units: units, Groups: groups, Active: active, ID: userId})
	}
	return users, nil
}

func AdminUpdateUser(user User) error {
	stmt, err := db.Prepare("UPDATE users SET active=$1 WHERE user_id=$2;")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(user.Active, user.ID)
	if err != nil {
		return err
	}
	//TODO update groups
	return nil
}

func UserUpdateUser(user User) error {
	//TODO find out wether active==emailVerified
	stmt, err := db.Prepare("UPDATE users SET active=$1, username=$2 WHERE user_id=$3;")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(user.Active, user.Username, user.ID)
	if err != nil {
		return err
	}
	return nil
}
