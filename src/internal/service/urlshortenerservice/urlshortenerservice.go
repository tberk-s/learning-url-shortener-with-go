package urlshortenerservice

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/db"
	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/urlshortenererror"
)

type URLShortenerService struct {
	db db.Database
}

func New(database db.Database) (*URLShortenerService, error) {
	return &URLShortenerService{
		db: database,
	}, nil
}

func (s *URLShortenerService) ShortenURL(originalURL string) (string, error) {
	// Validate URL format
	if !strings.HasPrefix(originalURL, "http://") && !strings.HasPrefix(originalURL, "https://") {
		originalURL = "https://" + originalURL
	}

	// Parse URL to validate format
	parsedURL, err := url.Parse(originalURL)
	if err != nil || parsedURL.Host == "" {
		return "", urlshortenererror.Wrap(
			err,
			"Invalid URL format. Example: example.org or https://example.org",
			http.StatusBadRequest,
			urlshortenererror.ErrInvalidInput,
		)
	}

	// Check if the host contains at least one dot (.) and has characters on both sides
	host := parsedURL.Host
	if !strings.Contains(host, ".") ||
		strings.HasPrefix(host, ".") ||
		strings.HasSuffix(host, ".") {
		return "", urlshortenererror.Wrap(
			nil,
			"Invalid domain format. URL must contain a valid domain (e.g., example.org)",
			http.StatusBadRequest,
			urlshortenererror.ErrInvalidInput,
		)
	}
	attempt := 0
	for {
		hash := sha256.New()
		log.Println(originalURL)
		hashInput := fmt.Sprintf("%s:%d", originalURL, attempt)
		hash.Write([]byte(hashInput))
		hashURL := hex.EncodeToString(hash.Sum(nil))
		log.Println(hashURL)
		log.Println(hashInput)
		shortURL := hashURL[:6]

		result, err := s.db.StoreURLs(shortURL, originalURL)
		if err != nil {
			var webErr *urlshortenererror.WebError
			if errors.As(err, &webErr) && webErr.ErrType == urlshortenererror.ErrDuplicate {
				attempt++
				continue
			}
			return "", err
		}
		return result, nil
	}
}
