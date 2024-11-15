package urlshortenerservice

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/db"
	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/urlshortenererror"
)

type URLShortenerService struct {
	counter uint64
	db      *db.DB
	mu      sync.Mutex
}

func New(db *db.DB) (*URLShortenerService, error) {
	// Get last counter from DB
	lastCounter, err := db.GetLastCounter()
	if err != nil {
		return nil, err
	}

	return &URLShortenerService{
		counter: lastCounter,
		db:      db,
	}, nil
}

func (s *URLShortenerService) ShortenURL(originalURL string) (string, error) {

	s.mu.Lock()
	defer s.mu.Unlock()
	// Atomically increment counter
	count := atomic.AddUint64(&s.counter, 1)

	hashInput := fmt.Sprintf("%s:%d", originalURL, count)
	hash := sha256.New()
	hash.Write([]byte(hashInput))
	hashURL := hex.EncodeToString(hash.Sum(nil))
	shortURL := hashURL[:6]

	// Store URL with counter
	if _, err := s.db.StoreURLs(shortURL, originalURL, count); err != nil {
		return "", fmt.Errorf("failed to store shortened URL: %w", err)
	}

	maxRetries := 3
	var err error

	for i := 0; i < maxRetries; i++ {
		_, err = s.db.StoreURLs(shortURL, originalURL, count)
		if err == nil {
			return shortURL, nil
		}

		var webErr *urlshortenererror.WebError
		if errors.As(err, &webErr) && errors.Is(webErr.ErrType, urlshortenererror.ErrDuplicate) {
			// If duplicate, generate new hash with updated counter
			count = atomic.AddUint64(&s.counter, 1)
			hashInput = fmt.Sprintf("%s:%d", originalURL, count)
			hash.Reset()
			hash.Write([]byte(hashInput))
			hashURL = hex.EncodeToString(hash.Sum(nil))
			shortURL = hashURL[:6]
			continue
		}

		return "", fmt.Errorf("failed to store shortened URL: %w", err)
	}

	return "", fmt.Errorf("failed to generate unique short URL after %d attempts: %w", maxRetries, err)
}
