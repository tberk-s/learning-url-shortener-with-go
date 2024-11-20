package urlshortenerservice_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/service/urlshortenerservice"
	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/urlshortenererror"
)

type MockDB struct {
	mock.Mock
}

func (m *MockDB) StoreURLs(shortURL, originalURL string) (string, error) {
	args := m.Called(shortURL, originalURL)
	return args.String(0), args.Error(1)
}

func (m *MockDB) GetOriginalURL(shortURL string) (string, error) {
	args := m.Called(shortURL)
	return args.String(0), args.Error(1)
}

func (m *MockDB) Close() {
	// Mock does not require implementation
}

func TestNew_Success(t *testing.T) {
	mockDB := new(MockDB)
	service, err := urlshortenerservice.New(mockDB)

	assert.NoError(t, err)
	assert.NotNil(t, service)
}

func TestShortenURL_Success(t *testing.T) {
	mockDB := new(MockDB)
	service, _ := urlshortenerservice.New(mockDB)

	originalURL := "https://example.org"
	shortURL := "abc123"

	mockDB.On("StoreURLs", mock.Anything, originalURL).Return(shortURL, nil).Once()

	result, err := service.ShortenURL(originalURL)

	assert.NoError(t, err)
	assert.Equal(t, shortURL, result)

	mockDB.AssertExpectations(t)
}

func TestShortenURL_InvalidURLFormat(t *testing.T) {
	mockDB := new(MockDB)
	service, _ := urlshortenerservice.New(mockDB)

	invalidURL := "://example"
	_, err := service.ShortenURL(invalidURL)

	assert.Error(t, err)

	var webErr *urlshortenererror.WebError
	if errors.As(err, &webErr) {
		assert.Equal(t, http.StatusBadRequest, webErr.Code)
		assert.Equal(t, urlshortenererror.ErrInvalidInput, webErr.ErrType)
	}
}

func TestShortenURL_InvalidDomain(t *testing.T) {
	mockDB := new(MockDB)
	service, _ := urlshortenerservice.New(mockDB)

	invalidDomainURL := "https://.example"
	_, err := service.ShortenURL(invalidDomainURL)

	assert.Error(t, err)

	var webErr *urlshortenererror.WebError
	if errors.As(err, &webErr) {
		assert.Equal(t, http.StatusBadRequest, webErr.Code)
		assert.Equal(t, urlshortenererror.ErrInvalidInput, webErr.ErrType)
	}
}

func TestShortenURL_DuplicateShortURL(t *testing.T) {
	mockDB := new(MockDB)
	service, _ := urlshortenerservice.New(mockDB)

	originalURL := "https://example.org"
	shortURL := "final123"

	// First attempt fails due to duplicate entry
	mockDB.On("StoreURLs", mock.Anything, originalURL).Return("", &urlshortenererror.WebError{
		ErrType:  urlshortenererror.ErrDuplicate, // Changed from ErrDuplicateEntry to ErrDuplicate
		Message:  "Duplicate short URL",
		Code:     http.StatusConflict,
		InnerErr: errors.New("unique violation"),
	}).Once()

	// Second attempt succeeds
	mockDB.On("StoreURLs", mock.Anything, originalURL).Return(shortURL, nil).Once()

	result, err := service.ShortenURL(originalURL)

	assert.NoError(t, err)
	assert.Equal(t, shortURL, result)

	mockDB.AssertExpectations(t)
}

func TestShortenURL_UnexpectedDBError(t *testing.T) {
	mockDB := new(MockDB)
	service, _ := urlshortenerservice.New(mockDB)

	originalURL := "https://example.org"

	mockDB.On("StoreURLs", mock.Anything, originalURL).Return("", errors.New("db connection error")).Once()

	_, err := service.ShortenURL(originalURL)

	assert.Error(t, err)
	assert.EqualError(t, err, "db connection error")

	mockDB.AssertExpectations(t)
}

func TestShortenURL_NoProtocolAdded(t *testing.T) {
	mockDB := new(MockDB)
	service, _ := urlshortenerservice.New(mockDB)

	originalURL := "example.org"
	shortURL := "xyz789"

	mockDB.On("StoreURLs", mock.Anything, "https://"+originalURL).Return(shortURL, nil).Once()

	result, err := service.ShortenURL(originalURL)

	assert.NoError(t, err)
	assert.Equal(t, shortURL, result)

	mockDB.AssertExpectations(t)
}

func TestShortenURL_HashCollision(t *testing.T) {
	mockDB := new(MockDB)
	service, _ := urlshortenerservice.New(mockDB)

	originalURL := "https://example.org"
	finalShortURL := "def456"

	// First attempt results in a collision
	mockDB.On("StoreURLs", mock.Anything, originalURL).
		Return("", &urlshortenererror.WebError{
			ErrType:  urlshortenererror.ErrDuplicate,
			Message:  "URL hash collision",
			Code:     http.StatusConflict,
			InnerErr: errors.New("unique violation"),
		}).Times(55)
	// Second attempt succeeds
	mockDB.On("StoreURLs", mock.Anything, originalURL).
		Return(finalShortURL, nil).Once()

	result, err := service.ShortenURL(originalURL)

	assert.NoError(t, err)
	assert.Equal(t, finalShortURL, result)
	mockDB.AssertExpectations(t)
}
