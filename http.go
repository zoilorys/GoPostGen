package main

import (
	"net/http"
	"html/template"
	"regexp"
	"os"
	"strings"
	"pages"
)

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

var templates = template.Must(template.ParseFiles("src/templates/notExist.html",
	"src/templates/view.html",
	"src/templates/edit.html",
	"src/templates/front.html",
	"src/templates/404.html"))

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.ListenAndServe(":8080", nil)
}

func makeHandler(f func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		matches := validPath.FindStringSubmatch(req.URL.Path)
		if matches == nil {
			http.NotFound(res, req)
			return
		}
		f(res, req, matches[2])
	}
}

func indexHandler(res http.ResponseWriter, req *http.Request) {
	dir, err := os.Open("pages")
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	names, rerr := dir.Readdirnames(-1)
	if rerr != nil {
		http.Error(res, rerr.Error(), http.StatusInternalServerError)
		return
	}
	parsedNames := names
	for i, name := range names {
		parsedNames[i] = strings.Split(name, ".")[0]
	}
	terr := templates.ExecuteTemplate(res, "front.html", struct{Names []string}{names})
	if terr != nil {
		http.Error(res, terr.Error(), http.StatusInternalServerError)
	}
}

func viewHandler(res http.ResponseWriter, req *http.Request, title string) {
	p, err := pages.LoadPage(title)
	if err != nil {
		renderTemplate(res, "notExist", &pages.Page{Title: title})
	} else {
		renderTemplate(res, "view", p)
	}
}

func editHandler(res http.ResponseWriter, req *http.Request, title string) {
	p, err := pages.LoadPage(title)
	if err != nil {
		p = &pages.Page{Title: title}
	}
	renderTemplate(res, "edit", p)
}

func saveHandler(res http.ResponseWriter, req *http.Request, title string) {
	body := req.FormValue("body")
	p := &pages.Page{Title: title, Body: []byte(body)}
	err := p.Save()
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(res, req, "/view/" + title, http.StatusFound)
}

func renderTemplate(res http.ResponseWriter, tmpl string, page *pages.Page) {
	err := templates.ExecuteTemplate(res, tmpl + ".html", page)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}