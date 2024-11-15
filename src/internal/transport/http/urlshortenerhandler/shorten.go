package urlshortenerhandler

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/db"
	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/service/urlshortenerservice"
)

type Handler struct {
	service *urlshortenerservice.URLShortenerService
	db      *db.DB
}

func New(db *db.DB) (*Handler, error) {
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

		if !strings.HasPrefix(originalURL, "http://") && !strings.HasPrefix(originalURL, "https://") {
			originalURL = "https://" + originalURL
		}

		shortURL, err := h.service.ShortenURL(originalURL)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to shorten URL: %v", err), http.StatusInternalServerError)
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
