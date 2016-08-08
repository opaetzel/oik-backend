package main

type Image struct {
	path    string
	Caption string `json:"caption"`
	Credits string `json:"credits"`
	ID      int    `json:"id"`
}

type RotateImage struct {
	Basepath string
	Num      int
	Caption  string `json:"caption"`
	Credits  string `json:"credits"`
	ID       int    `json:"id"`
}

type Text struct {
	Markdown string `json:"markdown"`
	Html     string `json:"html"`
	ID       int    `json:"id"`
}

type Column struct {
	Image *Image `json:"image"`
	Text  *Text  `json:"text"`
	ID    int    `json:"id"`
}

type Row struct {
	Left  Column `json:"left"`
	Right Column `json:"right"`
	ID    int    `json:"id"`
}

type Page struct {
	Title string `json:"title"`
	Rows  []Row  `json:"rows"`
	ID    int    `json:"id"`
}

type Unit struct {
	Title     string   `json:"title"`
	PageIds   []string `json:"pageids"`
	Published bool     `json:"published"`
	UserId    string   `json:"userid"`
	ID        int      `json:"id"`
}
