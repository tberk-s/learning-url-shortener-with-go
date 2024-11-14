package controllers

import (
	"net/http"
	"strings"

	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/db"
)

func RedirectHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortPath := r.URL.Path[1:]

		originalURL, err := database.GetOriginalURL(shortPath)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// Ensure URL has protocol
		if !strings.HasPrefix(originalURL, "http://") && !strings.HasPrefix(originalURL, "https://") {
			originalURL = "https://" + originalURL
		}

		http.Redirect(w, r, originalURL, http.StatusMovedPermanently)
	}
}
