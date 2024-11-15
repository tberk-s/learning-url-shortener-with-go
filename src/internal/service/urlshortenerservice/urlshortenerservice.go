package urlshortenerservice

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync/atomic"

	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/db"
)

type URLShortenerService struct {
	counter uint64
	db      *db.DB
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

	return shortURL, nil
}
