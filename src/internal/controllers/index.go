package controllers

import (
	"html/template"
	"net/http"
)

func ShowHomePage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("src/internal/views/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
