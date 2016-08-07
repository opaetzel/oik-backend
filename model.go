package main

type Image struct {
	path    string
	Caption string
	Credits string
	ID      int
}

type RotateImage struct {
	basepath string
	num      int
	Caption  string
	Credits  string
	ID       int
}

type Text struct {
	Markdown string
	Html     string
	ID       int
}

type Column struct {
	Image *Image
	Text  *Text
	ID    int
}

type Row struct {
	Left  Column
	Right Column
	ID    int
}

type Page struct {
	Title string
	Rows  []Row
	ID    int
}

type Unit struct {
	Title     string
	PageIds   []string
	Published bool
	UserId    string
	ID        int
}
