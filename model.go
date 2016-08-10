package main

import "encoding/json"

type Image struct {
	path    string
	Caption string `json:"caption" db:"caption"`
	Credits string `json:"credits" db:"credits"`
	ID      int    `json:"id" db:"id"`
}

type RotateImage struct {
	Basepath string `json:"basepath" db:"basepath"`
	Num      int    `json:"num" db:"num"`
	Caption  string `json:"caption" db:"caption"`
	Credits  string `json:"credits" db:"credits"`
	ID       int    `json:"id" db:"id"`
}

type Row struct {
	LeftMarkdown  string `json:"left_markdown" db:"left_markdown"`
	LeftHtml      string `json:"left_html" db:"left_html"`
	RightMarkdown string `json:"right_markdown" db:"right_markdown"`
	RightHtml     string `json:"right_html" db:"right_html"`
	ID            int    `json:"id" db:"row_id"`
}

type Page struct {
	Title  string `json:"title" db:"title"`
	Rows   []Row  `json:"rows" db:"rows"`
	UnitID int    `json:"unit_id" db:"unit_id"`
	ID     int    `json:"id" db:"id"`
}

type Unit struct {
	Title       string `json:"title" db:"title"`
	UnitImageID int    `json:"rotate_image_id" db:"rotate_image_id"`
	PageIds     []int  `json:"pageids" db:"pageids"`
	Published   bool   `json:"published" db:"published"`
	UserId      int    `json:"userid" db:"userid"`
	ID          int    `json:"id" db:"id"`
}

func (r *Image) UnmarshalJSON(data []byte) error {
	type Alias Image
	aux := &struct {
		MyID int `json:"image_id"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if r.ID == 0 {
		r.ID = aux.MyID
	}
	return nil
}

func (r *RotateImage) UnmarshalJSON(data []byte) error {
	type Alias RotateImage
	aux := &struct {
		MyID int `json:"rotate_image_id"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if r.ID == 0 {
		r.ID = aux.MyID
	}
	return nil
}

func (r *Row) UnmarshalJSON(data []byte) error {
	type Alias Row
	aux := &struct {
		MyID int `json:"row_id"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if r.ID == 0 {
		r.ID = aux.MyID
	}
	return nil
}

func (r *Page) UnmarshalJSON(data []byte) error {
	type Alias Page
	aux := &struct {
		MyID int `json:"page_id"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if r.ID == 0 {
		r.ID = aux.MyID
	}
	return nil
}

func (r *Unit) UnmarshalJSON(data []byte) error {
	type Alias Unit
	aux := &struct {
		MyID int `json:"unit_id"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if r.ID == 0 {
		r.ID = aux.MyID
	}
	return nil
}
