package pages

import (
	"io/ioutil"
)

type Page struct {
	Title string
	Body []byte
}

func (p *Page) Save() error {
	filename := "pages/" + p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func LoadPage(title string) (*Page, error) {
	filename := "pages/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{ Title: title, Body: body}, nil
}