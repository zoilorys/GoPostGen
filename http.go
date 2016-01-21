package main

import (
	"log"
	"net/http"
	"html/template"
	"regexp"
	"pages"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

var templates = template.Must(template.ParseFiles("src/templates/notExist.html",
	"src/templates/view.html",
	"src/templates/edit.html",
	"src/templates/front.html",
	"src/templates/404.html"))

var mongoURL = "mongodb://127.0.0.1:27017"

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
	mongo, merr := mgo.Dial(mongoURL)
	if merr != nil {
		panic(merr)
	}
	defer mongo.Close()

	var result []pages.Page
	collection := mongo.DB("gowebtest").C("pages")
	qerr := collection.Find(bson.M{}).All(&result)
	if qerr != nil {
		log.Fatal(qerr)
	}
	names := make([]string, len(result))
	for i, page := range result {
		names[i] = page.Title
	}
	terr := templates.ExecuteTemplate(res, "front.html", struct{Names []string}{names})
	if terr != nil {
		http.Error(res, terr.Error(), http.StatusInternalServerError)
	}
}


func viewHandler(res http.ResponseWriter, req *http.Request, title string) {
	mongo, merr := mgo.Dial(mongoURL)
	if merr != nil {
		panic(merr)
	}
	defer mongo.Close()

	page := pages.Page{}
	collection := mongo.DB("gowebtest").C("pages")

	err := collection.Find(bson.M{"title": title}).One(&page)
	if err != nil {
		renderTemplate(res, "notExist", &pages.Page{Title: title})
	} else {
		renderTemplate(res, "view", &page)
	}
}

func editHandler(res http.ResponseWriter, req *http.Request, title string) {
	mongo, merr := mgo.Dial(mongoURL)
	if merr != nil {
		panic(merr)
	}
	defer mongo.Close()

	page := pages.Page{}
	collection := mongo.DB("gowebtest").C("pages")

	err := collection.Find(bson.M{"title": title}).One(&page)
	if err != nil {
		page = pages.Page{Title: title}
	}
	renderTemplate(res, "edit", &page)
}

func saveHandler(res http.ResponseWriter, req *http.Request, title string) {
	mongo, merr := mgo.Dial(mongoURL)
	if merr != nil {
		panic(merr)
	}
	defer mongo.Close()

	body := req.FormValue("body")
	p := &pages.Page{Title: title, Body: []byte(body)}

	collection := mongo.DB("gowebtest").C("pages")

	_, err := collection.Upsert(bson.M{"title": p.Title}, p)
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