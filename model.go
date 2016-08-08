package main

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

type Text struct {
	Markdown string `json:"markdown" db:"markdown"`
	Html     string `json:"html" db:"html"`
	ID       int    `json:"id" db:"id"`
}

type Column struct {
	Image *Image `json:"image" db:"image"`
	Text  *Text  `json:"text" db:"text"`
	ID    int    `json:"id" db:"id"`
}

type Row struct {
	Left  Column `json:"left" db:"left"`
	Right Column `json:"right" db:"right"`
	ID    int    `json:"id" db:"id"`
}

type Page struct {
	Title string `json:"title" db:"title"`
	Rows  []Row  `json:"rows" db:"rows"`
	ID    int    `json:"id" db:"id"`
}

type Unit struct {
	Title     string       `json:"title" db:"title"`
	UnitImage *RotateImage `json:"unitimage" db:"unitimage"`
	PageIds   []int        `json:"pageids" db:"pageids"`
	Published bool         `json:"published" db:"published"`
	UserId    int          `json:"userid" db:"userid"`
	ID        int          `json:"id" db:"id"`
}
