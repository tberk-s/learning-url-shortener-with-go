package urlshortenerservice

import (
	"errors"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/db"
	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/urlshortenererror"
)

const (
	shortURLLength = 6 // Length of generated short URLs

	httpsPrefix = "https://"

	httpPrefix = "http://"

	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// URLShortenerService handles the business logic for URL shortening.
type URLShortenerService struct {
	db db.Database
}

// New creates a new URLShortenerService instance.
func New(database db.Database) (*URLShortenerService, error) {
	if database == nil {
		return nil, urlshortenererror.Wrap(

			nil,

			"Database instance cannot be nil",

			http.StatusInternalServerError,

			urlshortenererror.ErrServerError,
		)
	}

	return &URLShortenerService{db: database}, nil
}

// ShortenURL takes a URL and returns a shortened version.
func (s URLShortenerService) ShortenURL(originalURL string) (string, error) {
	if originalURL == "" {
		return "", urlshortenererror.Wrap(

			nil,

			"URL cannot be empty",

			http.StatusBadRequest,

			urlshortenererror.ErrInvalidInput,
		)
	}

	// Normalize URL

	originalURL = normalizeURL(originalURL)

	// Validate URL

	if err := validateURL(originalURL); err != nil {
		return "", err
	}

	// Generate short URL with collision handling

	return s.generateUniqueShortURL(originalURL)
}

func (s URLShortenerService) generateUniqueShortURL(originalURL string) (string, error) {
	var result string

	var err error

	for {
		shortURL := generateHash()

		result, err = s.db.StoreURLs(shortURL, originalURL)

		if err == nil {
			break
		}

		var webErr *urlshortenererror.WebError

		if !errors.As(err, &webErr) {
			return "", urlshortenererror.Wrap(err, "failed to store URL", http.StatusInternalServerError, urlshortenererror.ErrServerError)
		}

		if !errors.Is(webErr.ErrType, urlshortenererror.ErrDuplicate) {
			return "", webErr
		}

		log.Printf("Collision detected! Short URL %s already exists for original URL %s, trying again...", shortURL, originalURL)
	}

	return result, nil
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

var lock sync.Mutex

// generateHash creates a hash of the original URL with an attempt counter.
func generateHash() string {
	lock.Lock()

	defer lock.Unlock()

	result := make([]byte, shortURLLength)

	for i := range result {
		result[i] = charset[rng.Intn(len(charset))]
	}

	return string(result)
}

// normalizeURL ensures the URL has a proper protocol prefix.
func normalizeURL(urlStr string) string {
	if !strings.HasPrefix(urlStr, httpPrefix) && !strings.HasPrefix(urlStr, httpsPrefix) {
		return httpsPrefix + urlStr
	}

	return urlStr
}

// validateURL checks if the URL format is valid.
func validateURL(originalURL string) error {
	parsedURL, err := url.Parse(originalURL)

	if err != nil || parsedURL.Host == "" {
		return urlshortenererror.Wrap(

			err,

			"Invalid URL format. Example: example.org or https://example.org",

			http.StatusBadRequest,

			urlshortenererror.ErrInvalidInput,
		)
	}

	return validateHost(parsedURL.Host)
}

// validateHost ensures the domain name is properly formatted.
func validateHost(host string) error {
	if !strings.Contains(host, ".") ||

		strings.HasPrefix(host, ".") ||

		strings.HasSuffix(host, ".") {
		return urlshortenererror.Wrap(

			nil,

			"Invalid domain format. URL must contain a valid domain (e.g., example.org)",

			http.StatusBadRequest,

			urlshortenererror.ErrInvalidInput,
		)
	}

	return nil
}
