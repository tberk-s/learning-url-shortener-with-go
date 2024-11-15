package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
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
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping the db: %w", err)
	}
	return &DB{pool: pool}, nil
}

// GetLastCounter retrieves the last used counter value
func (db *DB) GetLastCounter() (uint64, error) {
	var counter uint64
	err := db.pool.QueryRow(context.Background(),
		"SELECT COALESCE(MAX(counter), 0) FROM urlmap").Scan(&counter)

	if err != nil {
		return 0, fmt.Errorf("failed to get last counter: %w", err)
	}
	return counter, nil
}

// UpdateCounter stores the new counter value with the URL
func (db *DB) StoreURLs(shortURL, originalURL string, counter uint64) (string, error) {
	_, err := db.pool.Exec(context.Background(),
		"INSERT INTO urlmap (short_url, original_url, counter) VALUES ($1, $2, $3)",
		shortURL, originalURL, counter)

	if err != nil {
		return "", fmt.Errorf("failed to store URL with counter: %w", err)
	}
	return shortURL, nil
}

func (db *DB) GetOriginalURL(shortURL string) (string, error) {
	var originalURL string
	err := db.pool.QueryRow(context.Background(), "SELECT original_url FROM urlmap WHERE short_url = $1", shortURL).Scan(&originalURL)
	return originalURL, err
}

func (db *DB) GetAllURLs() ([]URLMap, error) {
	var urls []URLMap
	err := db.pool.QueryRow(context.Background(), "SELECT * FROM urlmap").Scan(&urls)
	return urls, err
}

func (db *DB) DeleteURL(shortURL string) error {
	_, err := db.pool.Exec(context.Background(), "DELETE FROM urlmap WHERE short_url = $1", shortURL)
	return err
}

func (db *DB) Close() {
	db.pool.Close()
}
