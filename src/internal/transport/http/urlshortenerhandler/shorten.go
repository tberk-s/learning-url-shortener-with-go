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

type Handler struct {
	service *urlshortenerservice.URLShortenerService
	db      db.Database
}

func New(db db.Database) (*Handler, error) {
	service, err := urlshortenerservice.New(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL shortener service: %w", err)
	}

	return &Handler{
		service: service,
		db:      db,
	}, nil
}

func (h *Handler) ShowShortenPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		originalURL := r.FormValue("url")
		if originalURL == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}

		shortURL, err := h.service.ShortenURL(originalURL)
		if err != nil {
			var webErr *urlshortenererror.WebError
			if errors.As(err, &webErr) {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				log.Printf("Error: %v", webErr.Message) // Logs detailed error
				w.WriteHeader(webErr.Code)
				w.Write([]byte(webErr.Message)) // Sends detailed error message to client
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		// Success case
		tmpl, err := template.ParseFiles("src/internal/views/shorten.html")
		if err != nil {
			log.Printf("Template error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if err = tmpl.Execute(w, map[string]interface{}{
			"ShortURL": shortURL,
		}); err != nil {
			log.Printf("Template execution error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}
