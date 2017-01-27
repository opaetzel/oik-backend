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
	unit_id SERIAL PRIMARY KEY,
	front_image integer
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

CREATE TABLE IF NOT EXISTS cites (
	abbrev varchar(255),
	cite_text text,
	unit_id integer,
	cite_id SERIAL PRIMARY KEY
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
	image_id SERIAL PRIMARY KEY,
	age_known boolean default false,
	age integer default 0,
	imprecision integer default 0
);

CREATE TABLE IF NOT EXISTS users (
	username varchar(255),
	salt varchar(255),
	pwhash varchar(255),
	active boolean,
	user_id SERIAL PRIMARY KEY,
	mailhash varchar(255),
	points integer DEFAULT 0
);

CREATE TABLE IF NOT EXISTS groups (
	group_name varchar(255),
	group_id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS user_groups (
	user_id integer,
	group_id integer
);

CREATE TABLE IF NOT EXISTS row_results (
	decision varchar(30),
	row_id integer,
	page_result_id integer,
	row_result_id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS page_results (
	page_id integer,
	unit_id integer,
	user_id integer,
	page_result_id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS unit_results (
	pro_count smallint,
	con_count smallint,
	undecided_count smallint,
	unit_id integer,
	user_id integer
);

CREATE TABLE IF NOT EXISTS error_images (
	path varchar(255),
	correct_image_id integer,
	scale double precision,
	user_id integer,
	error_image_id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS error_circles (
	centerX integer,
	centerY integer,
	radius double precision,
	error_image_id integer,
	error_circle_id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS clicked_images (
	user_id integer,
	image_id integer
);

CREATE TABLE IF NOT EXISTS clicked_arguments (
	user_id integer,
	row_id integer
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
		var front_image int
		var unit_id int
		var pages_arr, images_arr, cites_arr string

		err := rows.Scan(&unit_title, &published, &rotate_image_id, &user_id, &color_scheme, &unit_id, &front_image, &pages_arr, &images_arr, &cites_arr)
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
		var cites []int
		if cites_arr == emptyArr {
			cites = make([]int, 0)
		} else {
			err = json.Unmarshal([]byte(cites_arr), &cites)
			if err != nil {
				return nil, err
			}
		}
		units = append(units, Unit{unit_title, rotate_image_id, pages, published, color_scheme, user_id, images, cites, front_image, unit_id})
	}
	return units, nil
}

func parseUnit(row *sql.Row) (Unit, error) {
	var unit_title string
	var published bool
	var rotate_image_id int
	var user_id int
	var color_scheme int
	var front_image int
	var unit_id int
	var pages_arr, images_arr, cites_arr string

	err := row.Scan(&unit_title, &published, &rotate_image_id, &user_id, &color_scheme, &unit_id, &front_image, &pages_arr, &images_arr, &cites_arr)
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
	var cites []int
	if cites_arr == emptyArr {
		cites = make([]int, 0)
	} else {
		err = json.Unmarshal([]byte(cites_arr), &cites)
		if err != nil {
			return Unit{}, err
		}
	}
	return Unit{unit_title, rotate_image_id, pages, published, color_scheme, user_id, images, cites, front_image, unit_id}, nil
}

func GetUnit(unitId int) (Unit, error) {
	row := db.QueryRow("SELECT units.*, json_agg(DISTINCT pages.page_id) AS pages_arr, json_agg(DISTINCT images.image_id) AS images_arr, json_agg(DISTINCT cites.cite_id) FROM units LEFT OUTER JOIN pages ON units.unit_id = pages.unit_id LEFT OUTER JOIN images ON units.unit_id=images.unit_id LEFT JOIN cites ON cites.unit_id=units.unit_id WHERE units.unit_id=$1 GROUP BY units.unit_id;", unitId)
	return parseUnit(row)
}

func GetAllUnits() ([]Unit, error) {
	rows, err := db.Query("SELECT units.*, json_agg(DISTINCT pages.page_id) AS pages_arr, json_agg(DISTINCT images.image_id) AS images_arr, json_agg(DISTINCT cites.cite_id FROM units LEFT OUTER JOIN pages ON units.unit_id = pages.unit_id LEFT OUTER JOIN images ON units.unit_id=images.unit_id LEFT JOIN cites ON cites.unit_id=units.unit_id GROUP BY units.unit_id;")
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	return parseUnits(rows)
}

/*
func GetUserUnits(userId int) ([]Unit, error) {
	rows, err := db.Query("SELECT units.*, json_agg(pages.page_id) AS pages_arr FROM units LEFT OUTER JOIN pages ON units.unit_id = pages.unit_id WHERE units.user_id=$1 GROUP BY units.unit_id;", userId)
defer rows.Close()
	if err != nil {
		return nil, err
	}
	return parseUnits(rows)
}
*/
func GetUnPublishedUnits() ([]Unit, error) {
	rows, err := db.Query("SELECT units.*, json_agg(DISTINCT pages.page_id) AS pages_arr, json_agg(DISTINCT images.image_id), json_agg(DISTINCT cites.cite_id)  FROM units LEFT OUTER JOIN pages ON units.unit_id = pages.unit_id LEFT OUTER JOIN images ON units.unit_id=images.unit_id LEFT JOIN cites ON cites.unit_id=units.unit_id WHERE units.published=false GROUP BY units.unit_id;")
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	return parseUnits(rows)
}

func GetPublishedUnits() ([]Unit, error) {
	rows, err := db.Query("SELECT units.*, json_agg(DISTINCT pages.page_id) AS pages_arr, json_agg(DISTINCT images.image_id), json_agg(DISTINCT cites.cite_id)  FROM units LEFT OUTER JOIN pages ON units.unit_id = pages.unit_id LEFT OUTER JOIN images ON units.unit_id=images.unit_id LEFT JOIN cites ON cites.unit_id=units.unit_id WHERE units.published=true GROUP BY units.unit_id;")
	defer rows.Close()
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

func DbUpdatePage(page Page) (Page, error) {
	stmt, err := db.Prepare("UPDATE pages SET page_title=$1, page_type=$2 WHERE page_id=$3;")
	if err != nil {
		return Page{}, err
	}
	_, err = stmt.Exec(page.Title, page.PageType, page.ID)
	if err != nil {
		return Page{}, err
	}
	stmt, err = db.Prepare("UPDATE rows SET left_markdown=$1, right_markdown=$2, left_has_image=$3, right_has_image=$4, leftimage=$5, rightimage=$6, left_is_argument=$7, right_is_argument=$8 WHERE row_id=$9 RETURNING row_id;")
	if err != nil {
		return Page{}, err
	}
	insStmt, err := db.Prepare("INSERT INTO rows (left_markdown, right_markdown, left_has_image, right_has_image, leftimage, rightimage, left_is_argument, right_is_argument, page_id) VALUES ($1, $2, $3, $4, $5 ,$6, $7, $8, $9) RETURNING row_id;")
	if err != nil {
		return Page{}, err
	}
	for idx, row := range page.Rows {
		dbRows, err := stmt.Query(row.LeftMarkdown, row.RightMarkdown, row.LeftHasImage, row.RightHasImage, row.LeftImage, row.RightImage, row.LeftIsArgument, row.RightIsArgument, row.ID)
		if err != nil {
			return Page{}, err
		}
		if !dbRows.Next() {
			var rowId int
			dbRow := insStmt.QueryRow(row.LeftMarkdown, row.RightMarkdown, row.LeftHasImage, row.RightHasImage, row.LeftImage, row.RightImage, row.LeftIsArgument, row.RightIsArgument, page.ID)
			dbRow.Scan(&rowId)
			if err != nil {
				return Page{}, err
			}
			log.Println("got new id for row", rowId)
			page.Rows[idx].ID = rowId
		}
		dbRows.Close()
	}
	return page, nil
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
	return Page{pageTitle, rows, unitId, page_type, pageId, userId, published, 0}, nil
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
		SELECT units.published, units.user_id, pages.page_title, pages.page_id, pages.unit_id, pages.page_type, json_agg(rows.* ORDER BY rows.row_id) AS rows FROM pages 
		LEFT JOIN rows ON rows.page_id = pages.page_id
		RIGHT JOIN units ON units.unit_id = pages.unit_id
		WHERE pages.page_id=$1
		GROUP BY pages.page_id, units.unit_id;
		`
	row := db.QueryRow(query, id)
	return parsePage(row)
}

func GetPageWithResultsById(pageId, userId int) (Page, error) {
	query := `
		SELECT units.published, units.user_id, pages.page_title, pages.unit_id, pages.page_type, json_agg(rows.* ORDER BY rows.row_id) AS rows, page_results.page_result_id FROM pages 
		LEFT JOIN rows ON rows.page_id = pages.page_id
		LEFT JOIN page_results ON page_results.page_id = pages.page_id AND page_results.user_id=$2
		RIGHT JOIN units ON units.unit_id = pages.unit_id
		WHERE pages.page_id=$1
		GROUP BY pages.page_id, units.unit_id, page_results.page_result_id;
		`
	row := db.QueryRow(query, pageId, userId)
	var published bool
	var pageTitle, page_type, jsonRows string
	var unitId, pageUserId int
	var pageResultId sql.NullInt64
	if err := row.Scan(&published, &pageUserId, &pageTitle, &unitId, &page_type, &jsonRows, &pageResultId); err != nil {
		return Page{}, err
	}
	var pageResultIdVal int
	if pageResultId.Valid {
		pageResultIdVal = int(pageResultId.Int64)
	}
	var rows []Row
	if jsonRows == emptyArr {
		rows = make([]Row, 0)
	} else {
		if err := json.Unmarshal([]byte(jsonRows), &rows); err != nil {
			return Page{}, err
		}
	}
	return Page{pageTitle, rows, unitId, page_type, pageId, pageUserId, published, pageResultIdVal}, nil
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
	row := db.QueryRow("INSERT INTO units (unit_title, published, rotate_image_id, user_id, color_scheme, front_image) VALUES ($1, $2, $3, $4, $5, $6) RETURNING units.unit_id", unit.Title, unit.Published, unit.UnitImageID, unit.UserId, unit.ColorScheme, unit.FrontImage)
	var id int
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func UpdateUnitAdmin(unit Unit) error {
	stmt, err := db.Prepare("UPDATE units SET unit_title=$1, published=$2, rotate_image_id=$3, color_scheme=$4, front_image=$5 WHERE unit_id=$6;")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(unit.Title, unit.Published, unit.UnitImageID, unit.ColorScheme, unit.FrontImage, unit.ID)
	if err != nil {
		return err
	}
	return nil
}

func UpdateUnitUser(unit Unit) error {
	stmt, err := db.Prepare("UPDATE units SET unit_title=$1, rotate_image_id=$2, color_scheme=$3, front_image=$4 WHERE unit_id=$5;")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(unit.Title, unit.UnitImageID, unit.ColorScheme, unit.FrontImage, unit.ID)
	if err != nil {
		return err
	}
	return nil
}

func UpdateErrorImage(errorImage ErrorImage) (ErrorImage, error) {
	stmt, err := db.Prepare("UPDATE error_images SET correct_image_id=$1, scale=$2 WHERE error_image_id=$3;")
	if err != nil {
		return ErrorImage{}, err
	}
	_, err = stmt.Exec(errorImage.CorrectImageId, errorImage.Scale, errorImage.ID)
	if err != nil {
		return ErrorImage{}, err
	}
	stmt, err = db.Prepare("DELETE FROM error_circles WHERE error_image_id=$1;")
	if err != nil {
		return ErrorImage{}, err
	}
	_, err = stmt.Exec(errorImage.ID)
	if err != nil {
		return ErrorImage{}, err
	}
	query := "INSERT INTO error_circles (radius, centerX, centerY, error_image_id) VALUES ($1,$2,$3,$4) RETURNING error_circle_id;"
	if err != nil {
		return ErrorImage{}, err
	}
	for idx, circle := range errorImage.ErrorCircles {
		var newId int
		err := db.QueryRow(query, circle.Radius, circle.CenterX, circle.CenterY, errorImage.ID).Scan(&newId)
		if err != nil {
			return ErrorImage{}, err
		}
		errorImage.ErrorCircles[idx].ID = newId
	}
	return errorImage, nil
}

func UpdateImageUser(image Image) error {
	stmt, err := db.Prepare("UPDATE images SET caption=$1, credits=$2, age_known=$3, age=$4, imprecision=$5 WHERE image_id=$6;")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(image.Caption, image.Credits, image.AgeKnown, image.Age, image.Imprecision, image.ID)
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
		SELECT units.published, units.user_id, images.path, images.caption, images.credits, images.unit_id, images.age_known, images.age, images.imprecision FROM units
		JOIN images ON images.unit_id = units.unit_id
		WHERE images.image_id = $1;
		`
	row := db.QueryRow(query, imageId)
	var published, ageKnown bool
	var userId, unitId, age, imprecision int
	var path, caption, credits string
	var nullPath sql.NullString
	err := row.Scan(&published, &userId, &nullPath, &caption, &credits, &unitId, &ageKnown, &age, &imprecision)
	if err != nil {
		return Image{}, err
	}
	if nullPath.Valid {
		path = nullPath.String
	}
	return Image{path, caption, credits, unitId, userId, imageId, published, ageKnown, age, imprecision}, nil
}

func GetRotateImageById(imageId int) (RotateImage, error) {
	query := `
		SELECT units.published, units.user_id, units.unit_id, rotate_images.basepath, rotate_images.caption, rotate_images.credits, rotate_images.num FROM units
		JOIN rotate_images ON rotate_images.rotate_image_id = units.rotate_image_id
		WHERE rotate_images.rotate_image_id = $1;
		`
	row := db.QueryRow(query, imageId)
	var published bool
	var userId, unitId, num int
	var path, caption, credits string
	var nullPath sql.NullString
	err := row.Scan(&published, &userId, &unitId, &nullPath, &caption, &credits, &num)
	if err != nil {
		return RotateImage{}, err
	}
	if nullPath.Valid {
		path = nullPath.String
	}
	return RotateImage{path, num, caption, credits, unitId, userId, imageId, published}, nil
}

func InsertImage(image Image) (int, error) {
	query := "INSERT INTO images (caption, credits, unit_id, age_known, age, imprecision) VALUES ($1, $2, $3, $4, $5, $6) RETURNING image_id;"
	var imageId int
	err := db.QueryRow(query, image.Caption, image.Credits, image.UnitId, image.AgeKnown, image.Age, image.Imprecision).Scan(&imageId)
	if err != nil {
		return -1, err
	}
	return int(imageId), nil
}

func InsertErrorImage(errorImage ErrorImage) (ErrorImage, error) {
	query := "INSERT INTO error_images (path, correct_image_id, scale, user_id) VALUES ($1, $2, $3, $4) RETURNING error_image_id;"
	var errorImageId int
	err := db.QueryRow(query, errorImage.path, errorImage.CorrectImageId, errorImage.Scale, errorImage.UserId).Scan(&errorImageId)
	if err != nil {
		return ErrorImage{}, err
	}
	errorImage.ID = errorImageId
	query = "INSERT INTO error_circles (radius, centerX, centerY, error_image_id) VALUES ($1,$2,$3,$4) RETURNING error_circle_id;"
	if err != nil {
		return ErrorImage{}, err
	}
	for idx, circle := range errorImage.ErrorCircles {
		var newId int
		err := db.QueryRow(query, circle.Radius, circle.CenterX, circle.CenterY, errorImage.ID).Scan(&newId)
		if err != nil {
			return ErrorImage{}, err
		}
		errorImage.ErrorCircles[idx].ID = newId
	}
	return errorImage, nil
}

func UpdateErrorImagePath(id int, path string) error {
	stmt, err := db.Prepare("UPDATE error_images SET path=$1 WHERE error_image_id=$2;")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(path, id)
	if err != nil {
		return err
	}
	return nil
}

func GetErrorImageById(id int) (ErrorImage, error) {
	query := "SELECT error_images.*, json_agg(error_circles.*) FROM error_images LEFT JOIN error_circles ON error_circles.error_image_id=error_images.error_image_id WHERE error_images.error_image_id=$1 GROUP BY error_images.error_image_id;"
	var path, circlesAgg string
	var scale float64
	var dbId, correctImageId, userId int
	err := db.QueryRow(query, id).Scan(&path, &correctImageId, &scale, &userId, &dbId, &circlesAgg)
	if err != nil {
		return ErrorImage{}, err
	}
	var errorCircles []Circle
	if err := json.Unmarshal([]byte(circlesAgg), &errorCircles); err != nil {
		return ErrorImage{}, err
	}
	return ErrorImage{path: path, CorrectImageId: correctImageId, Scale: scale, ID: dbId, ErrorCircles: errorCircles, UserId: userId}, nil
}

func InsertRotateImage(image RotateImage) (int, error) {
	query := "INSERT INTO rotate_images (caption, credits, num) VALUES ($1, $2, $3) RETURNING rotate_image_id;"
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
	query := `SELECT uo.username, uo.points, uo.active, 
		(
			SELECT COUNT(*) 
			FROM users ui
			WHERE (ui.points, ui.user_id) >= (uo.points, uo.user_id)
		) AS rank,
		json_agg(DISTINCT units.unit_id) AS units, 
		json_agg(DISTINCT groups.group_name) AS groups,
		json_agg(DISTINCT clicked_images.image_id) AS clicked_images,
		json_agg(DISTINCT clicked_arguments.row_id) AS clicked_arguments
		FROM users uo
		LEFT JOIN units ON units.user_id=uo.user_id 
		LEFT JOIN clicked_images ON clicked_images.user_id=uo.user_id 
		LEFT JOIN clicked_arguments ON clicked_arguments.user_id=uo.user_id 
		LEFT JOIN user_groups ON uo.user_id=user_groups.user_id 
		LEFT JOIN groups ON user_groups.group_id = groups.group_id
		WHERE uo.user_id=$1 GROUP BY uo.user_id`
	row := db.QueryRow(query, userId)
	var dbUsername, jsonUnits, jsonGroups, jsonClickedIms, jsonClickedArgs string
	var points, rank uint
	var active bool
	err := row.Scan(&dbUsername, &points, &active, &rank, &jsonUnits, &jsonGroups, &jsonClickedIms, &jsonClickedArgs)
	if err != nil {
		return User{}, err
	}
	var clickedIms []int
	if jsonClickedIms == emptyArr {
		clickedIms = make([]int, 0)
	} else {
		if err := json.Unmarshal([]byte(jsonClickedIms), &clickedIms); err != nil {
			return User{}, err
		}
	}
	var clickedArgs []int
	if jsonClickedArgs == emptyArr {
		clickedArgs = make([]int, 0)
	} else {
		if err := json.Unmarshal([]byte(jsonClickedArgs), &clickedArgs); err != nil {
			return User{}, err
		}
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
	u := User{Username: dbUsername, Units: units, Groups: groups, ID: userId, Points: points, Active: active, Rank: rank, ClickedImages: clickedIms, ClickedArguments: clickedArgs}
	return u, nil
}

func GetUserByName(username string) (User, error) {
	row := db.QueryRow("SELECT users.*, json_agg(groups.group_name) FROM users LEFT JOIN user_groups ON users.user_id=user_groups.user_id LEFT JOIN groups ON user_groups.group_id=groups.group_id WHERE username=$1 GROUP BY users.user_id", username)
	var dbUsername, salt, pwhash, jsonGroups, mailHash string
	var active bool
	var id int
	var points uint
	err := row.Scan(&dbUsername, &salt, &pwhash, &active, &id, &mailHash, &points, &jsonGroups)
	if err != nil {
		return User{}, err
	}
	var groups []string
	if err := json.Unmarshal([]byte(jsonGroups), &groups); err != nil {
		return User{}, err
	}
	u := User{dbUsername, groups, nil, id, salt, pwhash, active, mailHash, points, 0, "", nil, nil}
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
	query := `SELECT uo.username, uo.active, uo.user_id, uo.points, 
		json_agg(DISTINCT units.unit_id) AS units, 
		json_agg(DISTINCT groups.group_name),
		(
			SELECT COUNT(*) 
			FROM users ui
			WHERE (ui.points, ui.user_id) >= (uo.points, uo.user_id)
		) AS rank
		FROM users uo
		LEFT JOIN units ON units.user_id=uo.user_id 
		LEFT JOIN user_groups ON uo.user_id=user_groups.user_id 
		LEFT JOIN groups ON user_groups.group_id = groups.group_id
		GROUP BY uo.user_id`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := make([]User, 0)
	for rows.Next() {
		var dbUsername, jsonUnits, jsonGroups string
		var active bool
		var userId int
		var points, rank uint
		err := rows.Scan(&dbUsername, &active, &userId, &points, &jsonUnits, &jsonGroups, &rank)
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
		users = append(users, User{Username: dbUsername, Units: units, Groups: groups, Active: active, ID: userId, Points: points, Rank: rank})
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
	err = UpdateGroups(user)
	if err != nil {
		return err
	}
	return nil
}

func UpdateGroups(user User) error {
	//first delete all old groups for user
	stmt, err := db.Prepare("DELETE FROM user_groups WHERE user_id=$1;")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(user.ID)
	//now add groups
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err = tx.Prepare("INSERT INTO user_groups (user_id, group_id) VALUES ($1, (SELECT group_id FROM groups WHERE group_name=$2));")
	for _, group := range user.Groups {
		_, err := stmt.Exec(user.ID, group)
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

func UpdateUserPW(user User) error {
	stmt, err := db.Prepare("UPDATE users SET pwhash=$1 WHERE user_id=$2;")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(user.pwHash, user.ID)
	if err != nil {
		return err
	}
	return nil
}

func UserUpdateUser(user User) error {
	stmt, err := db.Prepare("UPDATE users SET active=$1, username=$2, points=$3 WHERE user_id=$4;")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(user.Active, user.Username, user.Points, user.ID)
	if err != nil {
		return err
	}
	//delete all clicked images of user
	stmt, err = db.Prepare("DELETE from clicked_images WHERE user_id=$1;")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(user.ID)
	if err != nil {
		return err
	}
	//delete all clicked arguments of user
	stmt, err = db.Prepare("DELETE from clicked_arguments WHERE user_id=$1;")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(user.ID)
	if err != nil {
		return err
	}
	//insert clicked images
	stmt, err = db.Prepare("INSERT INTO clicked_images (user_id, image_id) VALUES ($1, $2)")
	if err != nil {
		return err
	}
	for _, clickedIm := range user.ClickedImages {
		_, err = stmt.Exec(user.ID, clickedIm)
		if err != nil {
			return err
		}
	}
	//insert clicked arguments
	stmt, err = db.Prepare("INSERT INTO clicked_arguments (user_id, row_id) VALUES ($1, $2)")
	if err != nil {
		return err
	}
	for _, rowId := range user.ClickedArguments {
		_, err = stmt.Exec(user.ID, rowId)
		if err != nil {
			return err
		}
	}
	return nil
}

func SetRotateImagePathAndNum(imageId int, imageDir string, imageCount int) error {
	stmt, err := db.Prepare("UPDATE rotate_images SET basepath=$1, num=$2 WHERE rotate_image_id=$3")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(imageDir, imageCount, imageId)
	if err != nil {
		return err
	}
	return nil
}

func IsUsernameInDb(username string) (bool, error) {
	var count int
	row := db.QueryRow("SELECT COUNT(*) FROM users WHERE username=$1", username)
	err := row.Scan(&count)
	if err != nil {
		return true, err
	}
	return count > 0, nil
}

func GetRowOwnerId(rowId int) (int, error) {
	var userId int
	err := db.QueryRow("SELECT users.user_id FROM rows LEFT JOIN pages ON rows.page_id=pages.page_id LEFT JOIN units ON pages.unit_id=units.unit_id LEFT JOIN users ON units.user_id=users.user_id WHERE row_id=$1", rowId).Scan(&userId)
	if err != nil {
		return 0, err
	}
	return userId, nil
}

func RowDelete(rowId int) error {
	stmt, err := db.Prepare("DELETE FROM rows WHERE row_id=$1")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(rowId)
	if err != nil {
		return err
	}
	return nil
}

func DbUpdatePageResult(user User, pageResult PageResult) error {
	//TODO (maybe before calling this?) check wether page_result.user_id == user.id
	stmt, err := db.Prepare("UPDATE row_results SET decision=$1 WHERE page_result_id=$2 AND row_id=$3")
	if err != nil {
		return err
	}
	for _, rowResult := range pageResult.RowResults {
		_, err := stmt.Exec(rowResult.Decision, pageResult.Id, rowResult.RowID)
		if err != nil {
			return err
		}
	}
	return nil
}

func DbInsertPageResult(user User, pageResult PageResult) (int, error) {
	var pageResultId int
	err := db.QueryRow("INSERT INTO page_results (page_id, unit_id, user_id) VALUES ($1,$2,$3) RETURNING page_result_id", pageResult.PageId, pageResult.UnitId, user.ID).Scan(&pageResultId)
	if err != nil {
		return -1, err
	}
	stmt, err := db.Prepare("INSERT INTO row_results (decision, row_id, page_result_id) VALUES ($1,$2,$3)")
	if err != nil {
		return -1, err
	}
	for _, rowResult := range pageResult.RowResults {
		_, err := stmt.Exec(rowResult.Decision, rowResult.RowID, pageResultId)
		if err != nil {
			return -1, err
		}
	}
	return pageResultId, nil
}

func DbGetPageResult(pageResultId int) (PageResult, error) {
	var unitId, pageId, userId int
	var rowResultsAgg string
	err := db.QueryRow("SELECT page_results.unit_id, page_results.user_id, page_results.page_id, json_agg(row_results.* ORDER BY row_results.row_id) FROM page_results LEFT JOIN row_results ON row_results.page_result_id = page_results.page_result_id WHERE page_results.page_result_id=$1 GROUP BY page_results.page_result_id", pageResultId).Scan(&unitId, &userId, &pageId, &rowResultsAgg)
	if err != nil {
		return PageResult{}, err
	}
	var rowResults []Result
	err = json.Unmarshal([]byte(rowResultsAgg), &rowResults)
	if err != nil {
		return PageResult{}, err
	}
	return PageResult{rowResults, pageId, unitId, userId, pageResultId}, nil
}

func DbInsertUnitResult(user User, unitResult UnitResult) error {
	//TODO
	return nil
}

func DbUpdateUnitResult(user User, unitResult UnitResult) error {
	//TODO
	return nil
}

func DbGetUnitResults(unitId int) (UnitResult, error) {
	//TODO
	return UnitResult{}, nil
}

func DbDeleteUnit(unitId int) error {
	stmt, err := db.Prepare("DELETE FROM units WHERE unit_id=$1")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(unitId)
	if err != nil {
		return err
	}
	return nil
}

func DbDeletePage(pageId int) error {
	stmt, err := db.Prepare("DELETE FROM pages WHERE page_id=$1")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(pageId)
	if err != nil {
		return err
	}
	return nil
}

func GetErrorImages() ([]ErrorImage, error) {
	rows, err := db.Query("SELECT error_images.*, json_agg(error_circles.*) FROM error_images LEFT JOIN error_circles ON error_circles.error_image_id = error_images.error_image_id GROUP BY error_images.error_image_id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var path string
	var correctImageId int
	var scale float64
	var userId int
	var id int
	var errorCircleJson string
	imgs := make([]ErrorImage, 0)
	for rows.Next() {
		err := rows.Scan(&path, &correctImageId, &scale, &userId, &id, &errorCircleJson)
		if err != nil {
			return nil, err
		}
		var errorCircles []Circle
		err = json.Unmarshal([]byte(errorCircleJson), &errorCircles)
		if err != nil {
			return nil, err
		}
		imgs = append(imgs, ErrorImage{path, correctImageId, scale, errorCircles, userId, id})
	}
	return imgs, nil
}

/*
type PageResult struct {
	RowResults []Result `json:"rowResults`
	PageId     int      `json:"page"`
	UnitId     int      `json:"unit"`
	UserId     int      `json:"user"`
	Id         int      `json:"id"`
}
CREATE TABLE IF NOT EXISTS row_results (
	decision varchar(30),
	row_id integer,
	page_result_id integer
);

CREATE TABLE IF NOT EXISTS page_results (
	page_id integer,
	unit_id integer,
	user_id integer,
	page_result_id SERIAL PRIMARY KEY
);
func InsertCite(cite Cite) (int, error) {
	query := "INSERT INTO cites (abbrev, cite_text, unit_id) VALUES ($1, $2, $3) RETURNING cite_id;"
	var userId int
	err := db.QueryRow(query, cite.Abbrev, cite.Text, cite.UnitID).Scan(&userId)
	if err != nil {
		return -1, err
	}
	return userId, nil
}

func DeleteCite(cite Cite) error {
	stmt, err := db.Prepare("DELETE FROM cites WHERE cite_id=$1;")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(cite.ID)
	if err != nil {
		return err
	}
	return nil
}

func UpdateCite(cite Cite) error {
	stmt, err := db.Prepare("UPDATE cites SET abbrev=$1, cite_text=$2 WHERE cite_id=$3;")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(cite.Abbrev, cite.Text, cite.ID)
	if err != nil {
		return err
	}
	return nil
}
*/
