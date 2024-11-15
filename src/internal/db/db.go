package db

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/urlshortenererror"
)

type URLMap struct {
	ShortURL    string `db:"short_url"`
	OriginalURL string `db:"original_url"`
	Counter     uint64 `db:"counter"`
}

type DB struct {
	pool *pgxpool.Pool
}

func New(user, password, host, dbname string, port int) (*DB, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", user, password, host, port, dbname)

	pool, err := pgxpool.Connect(context.Background(), connStr)
	if err != nil {
		return nil, urlshortenererror.Wrap(err, "failed to connect to db", http.StatusInternalServerError, urlshortenererror.ErrDBConnection)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, urlshortenererror.Wrap(err, "failed to ping the db", http.StatusInternalServerError, urlshortenererror.ErrDBConnection)
	}
	return &DB{pool: pool}, nil
}

// GetLastCounter retrieves the last used counter value
func (db *DB) GetLastCounter() (uint64, error) {
	var counter uint64
	err := db.pool.QueryRow(context.Background(),
		"SELECT COALESCE(MAX(counter), 0) FROM urlmap").Scan(&counter)

	if err != nil {
		return 0, urlshortenererror.Wrap(err, "failed to get last counter", http.StatusInternalServerError, urlshortenererror.ErrDBQuery)
	}
	return counter, nil
}

// UpdateCounter stores the new counter value with the URL
func (db *DB) StoreURLs(shortURL, originalURL string, counter uint64) (string, error) {
	_, err := db.pool.Exec(context.Background(),
		"INSERT INTO urlmap (short_url, original_url, counter) VALUES ($1, $2, $3)",
		shortURL, originalURL, counter)

	if err != nil {
		return "", urlshortenererror.Wrap(err, "failed to store URL with counter", http.StatusInternalServerError, urlshortenererror.ErrDBQuery)
	}
	return shortURL, nil
}

func (db *DB) GetOriginalURL(shortURL string) (string, error) {
	var originalURL string
	err := db.pool.QueryRow(context.Background(), "SELECT original_url FROM urlmap WHERE short_url = $1", shortURL).Scan(&originalURL)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", urlshortenererror.Wrap(err, "URL not found", http.StatusNotFound, urlshortenererror.ErrNotFound)
		}
		return "", urlshortenererror.Wrap(err, "failed to get original URL", http.StatusInternalServerError, urlshortenererror.ErrDBQuery)
	}
	return originalURL, nil
}

// GetAllUrls ...
//	func (db *DB) GetAllURLs() ([]URLMap, error) {
//	var urls []URLMap
//	err := db.pool.QueryRow(context.Background(), "SELECT * FROM urlmap").Scan(&urls)
//	return urls, err
// }

// func (db *DB) DeleteURL(shortURL string) error {
// 	_, err := db.pool.Exec(context.Background(), "DELETE FROM urlmap WHERE short_url = $1", shortURL)
// 	return err
// 	}

func (db *DB) Close() {
	db.pool.Close()
}
