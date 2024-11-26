package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/urlshortenererror"
)

// Database interface to hold the database methods.
type Database interface {
	StoreURLs(shortURL, originalURL string) (string, error)
	GetOriginalURL(shortURL string) (string, error)
	Close()
}

// URLMap struct to hold the URL map.
type URLMap struct {
	CreatedAt   time.Time `db:"created_at"`
	ShortURL    string    `db:"short_url"`
	OriginalURL string    `db:"original_url"`
	Hits        int64     `db:"hits"`
}

// DB struct to hold the database connection pool.
type DB struct {
	pool *pgxpool.Pool
}

// New creates a new DB instance.
func New(user, password, host, dbname string, port int) (*DB, error) {
	config, err := pgxpool.ParseConfig(fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		user, password, host, port, dbname,
	))
	if err != nil {
		return nil, urlshortenererror.Wrap(err, "invalid connection config", http.StatusInternalServerError, urlshortenererror.ErrDBConnection)
	}

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, urlshortenererror.Wrap(err, "failed to connect to db", http.StatusInternalServerError, urlshortenererror.ErrDBConnection)
	}

	if pingErr := pool.Ping(context.Background()); pingErr != nil {
		return nil, urlshortenererror.Wrap(pingErr, "failed to ping the db", http.StatusInternalServerError, urlshortenererror.ErrDBConnection)
	}

	return &DB{pool: pool}, nil
}

// StoreURLs stores the short URL and original URL in the database.
func (db *DB) StoreURLs(shortURL, originalURL string) (string, error) {
	log.Printf("Attempting to store URL: short=%s, original=%s", shortURL, originalURL)

	tx, err := db.pool.Begin(context.Background())
	if err != nil {
		return "", urlshortenererror.Wrap(err, "failed to begin transaction", http.StatusInternalServerError, urlshortenererror.ErrDBQuery)
	}
	defer func() {
		if deferErr := tx.Rollback(context.Background()); deferErr != nil && !errors.Is(deferErr, pgx.ErrTxClosed) {
			log.Printf("Failed to rollback transaction: %v", deferErr)
		}
	}()

	var resultShortURL string

	// Try to update existing row and return in one query.
	err = tx.QueryRow(context.Background(),
		`UPDATE urlmap 
         SET hits = hits + 1
         WHERE original_url = $1
         RETURNING short_url`, // Removed the extra comma after hits + 1
		originalURL).Scan(&resultShortURL)

	if err == nil {
		return commitAndReturn(tx, resultShortURL)
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return "", urlshortenererror.Wrap(err, "failed to update URL", http.StatusInternalServerError, urlshortenererror.ErrDBQuery)
	}

	// Try to insert new row
	err = tx.QueryRow(context.Background(),
		`INSERT INTO urlmap (short_url, original_url, hits) 
         VALUES ($1, $2, 1) 
         RETURNING short_url`,
		shortURL, originalURL).Scan(&resultShortURL)

	if err == nil {
		return commitAndReturn(tx, resultShortURL)
	}

	// Handle insert errors
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		log.Println("URL HASH COLLISION", err, pgErr.Code, pgErr.Message)

		return "", urlshortenererror.Wrap(err, "URL hash collision", http.StatusConflict, urlshortenererror.ErrDuplicate)
	}

	return "", urlshortenererror.Wrap(err, "failed to insert URL", http.StatusInternalServerError, urlshortenererror.ErrDBQuery)
}

// Helper function to avoid repetition.
func commitAndReturn(tx pgx.Tx, shortURL string) (string, error) {
	if err := tx.Commit(context.Background()); err != nil {
		return "", urlshortenererror.Wrap(err, "failed to commit transaction", http.StatusInternalServerError, urlshortenererror.ErrDBQuery)
	}
	log.Printf("Successfully stored URL: %s", shortURL)

	return shortURL, nil
}

// GetOriginalURL gets the original URL from the short URL.
func (db *DB) GetOriginalURL(shortURL string) (string, error) {
	tx, err := db.pool.Begin(context.Background())
	if err != nil {
		return "", urlshortenererror.Wrap(err, "failed to begin transaction", http.StatusInternalServerError, urlshortenererror.ErrDBQuery)
	}
	defer func() {
		if deferErr := tx.Rollback(context.Background()); deferErr != nil && !errors.Is(deferErr, pgx.ErrTxClosed) {
			log.Printf("Failed to rollback transaction: %v", deferErr)
		}
	}()

	var originalURL string
	err = tx.QueryRow(context.Background(),
		`UPDATE urlmap 
         SET hits = hits + 1 
         WHERE short_url = $1 
         RETURNING original_url`,
		shortURL).Scan(&originalURL)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", urlshortenererror.Wrap(err, "URL not found", http.StatusNotFound, urlshortenererror.ErrNotFound)
		}

		return "", urlshortenererror.Wrap(err, "failed to get original URL", http.StatusInternalServerError, urlshortenererror.ErrDBQuery)
	}

	if err = tx.Commit(context.Background()); err != nil {
		return "", urlshortenererror.Wrap(err, "failed to commit transaction", http.StatusInternalServerError, urlshortenererror.ErrDBQuery)
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

// Close closes the database connection pool.
func (db *DB) Close() {
	db.pool.Close()
}
