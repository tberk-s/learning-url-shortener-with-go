package urlshortenerhandler

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/db"
	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/service/urlshortenerservice"
	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/urlshortenererror"
)

// Handler struct to hold the dependencies.
type Handler struct {
	service *urlshortenerservice.URLShortenerService
	db      *db.DB
}

// New creates a new Handler instance.
func New(database *db.DB) (*Handler, error) {
	service, err := urlshortenerservice.New(database)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL shortener service: %w", err)
	}

	return &Handler{
		service: service,
		db:      database,
	}, nil
}

// ShowShortenPage handles the request to show the shorten page.
func (h *Handler) ShowShortenPage() http.HandlerFunc {
	return func(wr http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(wr, "Method not allowed", http.StatusMethodNotAllowed)

			return
		}

		originalURL := req.FormValue("url")
		if originalURL == "" {
			http.Error(wr, "URL is required", http.StatusBadRequest)

			return
		}

		shortURL, err := h.service.ShortenURL(originalURL)
		if err != nil {
			var webErr *urlshortenererror.WebError
			if errors.As(err, &webErr) {
				wr.Header().Set("Content-Type", "text/plain; charset=utf-8")
				log.Printf("Error: %v", webErr.Message) // Logs detailed error
				wr.WriteHeader(webErr.Code)
				if _, writeErr := wr.Write([]byte(webErr.Message)); writeErr != nil {
					log.Printf("Failed to write error message: %v", writeErr)
				}

				return
			}
			http.Error(wr, "Internal server error", http.StatusInternalServerError)

			return
		}

		// Success case
		tmpl, err := template.ParseFiles("src/internal/views/shorten.html")
		if err != nil {
			log.Printf("Template error: %v", err)
			http.Error(wr, "Internal server error", http.StatusInternalServerError)

			return
		}

		if err = tmpl.Execute(wr, map[string]any{
			"ShortURL": shortURL,
		}); err != nil {
			log.Printf("Template execution error: %v", err)
			http.Error(wr, "Internal server error", http.StatusInternalServerError)

			return
		}
	}
}
