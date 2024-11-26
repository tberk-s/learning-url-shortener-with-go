package db_test

import (
	"errors"
	"testing"

	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/db"
	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/urlshortenererror"
)

// TestConfig holds test database configuration
var testConfig = struct {
	user     string
	password string
	host     string
	dbname   string
	port     int
}{
	user:     "test_user",
	password: "test_password",
	host:     "localhost",
	dbname:   "test_urlshortener",
	port:     5432,
}

func setupTestDB(t *testing.T) *db.DB {
	database, err := db.New(
		testConfig.user,
		testConfig.password,
		testConfig.host,
		testConfig.dbname,
		testConfig.port,
	)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	return database
}

func cleanupTestDB(database *db.DB) {
	if database != nil {
		database.Close()
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		user        string
		password    string
		host        string
		dbname      string
		port        int
		expectError bool
	}{
		{
			name:        "Valid configuration",
			user:        testConfig.user,
			password:    testConfig.password,
			host:        testConfig.host,
			dbname:      testConfig.dbname,
			port:        testConfig.port,
			expectError: false,
		},
		{
			name:        "Invalid host",
			user:        testConfig.user,
			password:    testConfig.password,
			host:        "invalid-host",
			dbname:      testConfig.dbname,
			port:        testConfig.port,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database, err := db.New(tt.user, tt.password, tt.host, tt.dbname, tt.port)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if database == nil {
					t.Error("Expected database instance but got nil")
				}
				if database != nil {
					database.Close()
				}
			}
		})
	}
}

func TestStoreURLs(t *testing.T) {
	database := setupTestDB(t)
	defer cleanupTestDB(database)

	tests := []struct {
		expectedErrType error
		name            string
		shortURL        string
		originalURL     string
		expectedError   bool
	}{
		{
			name:          "Store new URL",
			shortURL:      "abc123",
			originalURL:   "https://example.com",
			expectedError: false,
		},
		{
			name:            "Duplicate short URL",
			shortURL:        "abc123",
			originalURL:     "https://different.com",
			expectedError:   true,
			expectedErrType: urlshortenererror.ErrDuplicate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := database.StoreURLs(tt.shortURL, tt.originalURL)

			if tt.expectedError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				var webErr *urlshortenererror.WebError
				if !errors.As(err, &webErr) {
					t.Errorf("Expected WebError but got different error type: %v", err)
					return
				}
				if webErr.ErrType != tt.expectedErrType {
					t.Errorf("Expected error type %v but got %v", tt.expectedErrType, webErr.ErrType)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if result != tt.shortURL {
					t.Errorf("Expected short URL %s but got %s", tt.shortURL, result)
				}
			}
		})
	}
}

func TestGetOriginalURL(t *testing.T) {
	database := setupTestDB(t)
	defer cleanupTestDB(database)
	shortURL := "abc123"
	originalURL := "https://example.com"

	tests := []struct {
		expectedErrType error
		name            string
		shortURL        string
		expectedURL     string
		expectedError   bool
	}{
		{
			name:          "Get existing URL",
			shortURL:      shortURL,
			expectedURL:   originalURL,
			expectedError: false,
		},
		{
			name:            "Get non-existent URL",
			shortURL:        "nonexistent",
			expectedError:   true,
			expectedErrType: urlshortenererror.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := database.GetOriginalURL(tt.shortURL)

			if tt.expectedError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				var webErr *urlshortenererror.WebError
				if !errors.As(err, &webErr) {
					t.Errorf("Expected WebError but got different error type: %v", err)
					return
				}
				if webErr.ErrType != tt.expectedErrType {
					t.Errorf("Expected error type %v but got %v", tt.expectedErrType, webErr.ErrType)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if result != tt.expectedURL {
					t.Errorf("Expected URL %s but got %s", tt.expectedURL, result)
				}
			}
		})
	}
}

func TestConcurrentAccess(t *testing.T) {
	database := setupTestDB(t)
	defer cleanupTestDB(database)

	shortURL := "conc123"
	originalURL := "https://example123.com"

	// Store initial URL
	_, err := database.StoreURLs(shortURL, originalURL)
	if err != nil {
		t.Fatalf("Failed to store initial URL: %v", err)
	}

	// Run concurrent gets
	concurrentRequests := 10
	done := make(chan bool)

	for i := 0; i < concurrentRequests; i++ {
		go func() {
			_, err := database.GetOriginalURL(shortURL)
			if err != nil {
				t.Errorf("Concurrent get failed: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < concurrentRequests; i++ {
		<-done
	}
}
