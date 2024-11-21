package urlshortenerservice_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/service/urlshortenerservice"
	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/urlshortenererror"
)

// MockDB implements the Database interface for testing
type MockDB struct {
	storeURLsFunc func(shortURL, originalURL string) (string, error)
	getURLFunc    func(shortURL string) (string, error)
}

func (m *MockDB) StoreURLs(shortURL, originalURL string) (string, error) {
	return m.storeURLsFunc(shortURL, originalURL)
}

func (m *MockDB) GetOriginalURL(shortURL string) (string, error) {
	return m.getURLFunc(shortURL)
}

func (m *MockDB) Close() {}

func TestNew_Success(t *testing.T) {
	mockDB := &MockDB{}
	service, err := urlshortenerservice.New(mockDB)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if service == nil {
		t.Error("Expected service to not be nil")
	}
}

func TestShortenURL_Success(t *testing.T) {
	originalURL := "https://example.org"
	expectedShortURL := "abc123"

	mockDB := &MockDB{
		storeURLsFunc: func(shortURL, origURL string) (string, error) {
			if origURL != originalURL {
				t.Errorf("Expected originalURL %s, got %s", originalURL, origURL)
			}
			return expectedShortURL, nil
		},
	}

	service, _ := urlshortenerservice.New(mockDB)
	result, err := service.ShortenURL(originalURL)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != expectedShortURL {
		t.Errorf("Expected shortURL %s, got %s", expectedShortURL, result)
	}
}

func TestShortenURL_InvalidURLFormat(t *testing.T) {
	mockDB := &MockDB{}
	service, _ := urlshortenerservice.New(mockDB)

	invalidURL := "://example"
	_, err := service.ShortenURL(invalidURL)

	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
		return
	}

	var webErr *urlshortenererror.WebError
	if !errors.As(err, &webErr) {
		t.Error("Expected WebError type")
		return
	}

	if webErr.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, webErr.Code)
	}
	if webErr.ErrType != urlshortenererror.ErrInvalidInput {
		t.Errorf("Expected error type %v, got %v", urlshortenererror.ErrInvalidInput, webErr.ErrType)
	}
}

func TestShortenURL_DuplicateShortURL(t *testing.T) {
	originalURL := "https://example.org"
	finalShortURL := "final123"
	callCount := 0

	mockDB := &MockDB{
		storeURLsFunc: func(shortURL, origURL string) (string, error) {
			callCount++
			if callCount == 1 {
				return "", &urlshortenererror.WebError{
					ErrType:  urlshortenererror.ErrDuplicate,
					Message:  "Duplicate short URL",
					Code:     http.StatusConflict,
					InnerErr: errors.New("unique violation"),
				}
			}
			return finalShortURL, nil
		},
	}

	service, _ := urlshortenerservice.New(mockDB)
	result, err := service.ShortenURL(originalURL)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != finalShortURL {
		t.Errorf("Expected shortURL %s, got %s", finalShortURL, result)
	}
	if callCount != 2 {
		t.Errorf("Expected 2 calls to StoreURLs, got %d", callCount)
	}
}

func TestShortenURL_RandomCollision(t *testing.T) {
	originalURL := "https://google.com"
	duplicateKey1 := "abc123"
	duplicateKey2 := "def456"
	uniqueKey := "xyz789"
	callCount := 0

	mockDB := &MockDB{
		storeURLsFunc: func(shortURL, origURL string) (string, error) {
			callCount++
			switch callCount {
			case 1:
				// First attempt - returns first duplicate key error
				return duplicateKey1, &urlshortenererror.WebError{
					ErrType:  urlshortenererror.ErrDuplicate,
					Message:  "Duplicate short URL",
					Code:     http.StatusConflict,
					InnerErr: errors.New("unique violation"),
				}
			case 2:
				// Second attempt - returns second duplicate key error
				return duplicateKey2, &urlshortenererror.WebError{
					ErrType:  urlshortenererror.ErrDuplicate,
					Message:  "Duplicate short URL",
					Code:     http.StatusConflict,
					InnerErr: errors.New("unique violation"),
				}
			default:
				// Third attempt - finally succeeds
				return uniqueKey, nil
			}
		},
	}

	service, _ := urlshortenerservice.New(mockDB)

	result, err := service.ShortenURL(originalURL)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result == duplicateKey1 || result == duplicateKey2 {
		t.Errorf("Got duplicate key %s, expected a unique key", result)
	}
	if result != uniqueKey {
		t.Errorf("Expected final unique key %s, got %s", uniqueKey, result)
	}
}
