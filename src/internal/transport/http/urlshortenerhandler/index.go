package urlshortenerhandler

import (
	"html/template"
	"log"
	"net/http"
)

var indexTemplate = template.Must(template.ParseFiles("src/internal/views/index.html"))

// ShowHomePage handles the request to show the home page.
func ShowHomePage(wr http.ResponseWriter, _ *http.Request) {
	if err := indexTemplate.Execute(wr, nil); err != nil {
		log.Println("Template execution error:", err)
		http.Error(wr, err.Error(), http.StatusInternalServerError)
	}
}
