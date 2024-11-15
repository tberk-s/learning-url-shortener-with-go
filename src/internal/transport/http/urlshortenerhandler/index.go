package urlshortenerhandler

import (
	"html/template"
	"log"
	"net/http"
)

var indexTemplate = template.Must(template.ParseFiles("src/internal/views/index.html"))

func ShowHomePage(w http.ResponseWriter, r *http.Request) {
	if err := indexTemplate.Execute(w, nil); err != nil {
		log.Println("Template execution error:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
