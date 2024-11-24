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

type Database interface {
	StoreURLs(shortURL, originalURL string) (string, error)
	GetOriginalURL(shortURL string) (string, error)
	Close()
}

type URLMap struct {
	ShortURL    string    `db:"short_url"`
	OriginalURL string    `db:"original_url"`
	Hits        int64     `db:"hits"`
	CreatedAt   time.Time `db:"created_at"`
}

type DB struct {
	pool *pgxpool.Pool
}

func New(user, password, host, dbname string, port int) (*DB, error) {
	config, err := pgxpool.ParseConfig(fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		user, password, host, port, dbname,
	))
	if err != nil {
		return nil, urlshortenererror.Wrap(err, "invalid connection config", http.StatusInternalServerError, urlshortenererror.ErrDBConnection)
	}

	// Add pool configuration
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, urlshortenererror.Wrap(err, "failed to connect to db", http.StatusInternalServerError, urlshortenererror.ErrDBConnection)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, urlshortenererror.Wrap(err, "failed to ping the db", http.StatusInternalServerError, urlshortenererror.ErrDBConnection)
	}
	return &DB{pool: pool}, nil
}

func (db *DB) StoreURLs(shortURL, originalURL string) (string, error) {
	log.Printf("Attempting to store URL: short=%s, original=%s", shortURL, originalURL)

	tx, err := db.pool.Begin(context.Background())
	if err != nil {
		return "", urlshortenererror.Wrap(err, "failed to begin transaction", http.StatusInternalServerError, urlshortenererror.ErrDBQuery)
	}
	defer tx.Rollback(context.Background())

	var resultShortURL string

	// Try to update existing row and return in one query
	err = tx.QueryRow(context.Background(),
		`UPDATE urlmap 
         SET hits = hits + 1
         WHERE original_url = $1
         RETURNING short_url`, // Removed the extra comma after hits + 1
		originalURL).Scan(&resultShortURL)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No existing row, try to insert
			err = tx.QueryRow(context.Background(),
				`INSERT INTO urlmap (short_url, original_url, hits) 
                 VALUES ($1, $2, 1) 
                 RETURNING short_url`,
				shortURL, originalURL).Scan(&resultShortURL)

			if err != nil {
				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
					log.Println("URL HASH COLLISION", err, pgErr.Code, pgErr.Message)
					return "", urlshortenererror.Wrap(err, "URL hash collision", http.StatusConflict, urlshortenererror.ErrDuplicate)
				}
				return "", urlshortenererror.Wrap(err, "failed to insert URL", http.StatusInternalServerError, urlshortenererror.ErrDBQuery)
			}
		} else {
			return "", urlshortenererror.Wrap(err, "failed to update URL", http.StatusInternalServerError, urlshortenererror.ErrDBQuery)
		}
	}

	if err = tx.Commit(context.Background()); err != nil {
		return "", urlshortenererror.Wrap(err, "failed to commit transaction", http.StatusInternalServerError, urlshortenererror.ErrDBQuery)
	}

	log.Printf("Successfully stored URL: %s", resultShortURL)
	return resultShortURL, nil
}

func (db *DB) GetOriginalURL(shortURL string) (string, error) {
	tx, err := db.pool.Begin(context.Background())
	if err != nil {
		return "", urlshortenererror.Wrap(err, "failed to begin transaction", http.StatusInternalServerError, urlshortenererror.ErrDBQuery)
	}
	defer tx.Rollback(context.Background())

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

func (db *DB) Close() {
	db.pool.Close()
}
