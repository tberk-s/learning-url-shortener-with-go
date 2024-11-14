package controllers

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/db"
	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/shortener"
)

func ShowShortenPage(db *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		shortURL := shortener.ShortenURL(originalURL)
		time.Sleep(5 * time.Second)
		if _, err := db.StoreURLs(shortURL, originalURL); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := map[string]string{
			"ShortURL": shortURL,
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
}
