package controllers

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/shortener"
)

func ShowShortenPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		url := r.FormValue("url")
		fmt.Println(url)
		return
	}

	originalURL := r.FormValue("url")
	if originalURL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(originalURL, "http://") && !strings.HasPrefix(originalURL, "https://") {
		originalURL = "https://" + originalURL
	}

	data := map[string]string{
		"ShortURL": shortener.ShortenURL(originalURL),
	}

	tmpl, err := template.ParseFiles("src/internal/views/shorten.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
