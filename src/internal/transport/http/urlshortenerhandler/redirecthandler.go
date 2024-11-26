package urlshortenerhandler

import (
	"net/http"

	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/db"
)

// RedirectHandler handles the request to redirect to the original URL.
func RedirectHandler(database *db.DB) http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		shortPath := req.URL.Path[1:]

		if shortPath == "" {
			http.Error(wr, "URL not provided", http.StatusBadRequest)

			return
		}

		originalURL, err := database.GetOriginalURL(shortPath)
		if err != nil {
			http.Error(wr, err.Error(), http.StatusNotFound)

			return
		}

		http.Redirect(wr, req, originalURL, http.StatusPermanentRedirect)
	}
}
