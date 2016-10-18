package main

import "encoding/json"

type Image struct {
	path      string
	Caption   string `json:"caption" db:"caption"`
	Credits   string `json:"credits" db:"credits"`
	UnitId    int    `json:"unit" db:"unit_id"`
	UserId    int    `json:"user_id" db:"user_id"`
	ID        int    `json:"id" db:"id"`
	published bool
}

type RotateImage struct {
	basepath  string `json:"basepath" db:"basepath"`
	Num       int    `json:"numImages" db:"num"`
	Caption   string `json:"caption" db:"caption"`
	Credits   string `json:"credits" db:"credits"`
	UnitId    int    `json:"unit" db:"unit_id"`
	UserId    int    `json:"user_id" db:"user_id"`
	ID        int    `json:"id" db:"id"`
	published bool
}

type Row struct {
	LeftMarkdown    string `json:"left_markdown" db:"left_markdown"`
	RightMarkdown   string `json:"right_markdown" db:"right_markdown"`
	LeftHasImage    bool   `json:"left_has_image"`
	RightHasImage   bool   `json:"right_has_image"`
	LeftImage       int    `json:"leftImage,omitempty" db:"leftImage"`
	RightImage      int    `json:"rightImage,omitempty" db:"rightImage"`
	LeftIsArgument  bool   `json:"left_is_argument"`
	RightIsArgument bool   `json:"right_is_argument"`
	ID              int    `json:"id" db:"row_id"`
}

type Page struct {
	Title     string `json:"title" db:"title"`
	Rows      []Row  `json:"rows" db:"rows"`
	UnitID    int    `json:"unit" db:"unit_id"`
	PageType  string `json:"page_type" db:"page_type"`
	ID        int    `json:"id" db:"id"`
	userId    int
	published bool
}

type Unit struct {
	Title       string `json:"title" db:"title"`
	UnitImageID int    `json:"rotateImage" db:"rotate_image_id"`
	PageIds     []int  `json:"pages" db:"pageids"`
	Published   bool   `json:"published" db:"published"`
	ColorScheme int    `json:"color_scheme" db:"color_scheme"`
	UserId      int    `json:"user" db:"userid"`
	ImageIds    []int  `json:"images" db:"image_ids"`
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
		MyID    int   `json:"unit_id"`
		MyPages []int `json:"pageIds"`
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
	r.PageIds = aux.MyPages
	return nil
}
