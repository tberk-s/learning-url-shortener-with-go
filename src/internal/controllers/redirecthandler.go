package controllers

import (
	"net/http"
	"strings"

	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/db"
)

func RedirectHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortPath := r.URL.Path[1:]

		if shortPath == "" {
			http.Error(w, "URL not provided", http.StatusBadRequest)
			return
		}

		originalURL, err := database.GetOriginalURL(shortPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if !strings.HasPrefix(originalURL, "http://") && !strings.HasPrefix(originalURL, "https://") {
			originalURL = "https://" + originalURL
		}

		http.Redirect(w, r, originalURL, http.StatusPermanentRedirect)
	}
}
