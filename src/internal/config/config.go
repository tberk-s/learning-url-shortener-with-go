package config

import (
	"net/http"
	"os"
	"strconv"

	"github.com/tberk-s/learning-url-shortener-with-go/src/internal/urlshortenererror"
)

// Config struct to hold the configuration.
type Config struct {
	ServerEnv  string
	DBName     string
	DBHost     string
	DBUser     string
	DBPassword string
	DBPort     int
}

// LoadConfig loads the configuration from the environment variables.
func LoadConfig() (*Config, error) {
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		return nil, urlshortenererror.Wrap(err, "invalid db port", http.StatusInternalServerError, urlshortenererror.ErrInvalidDBPort)
	}

	return &Config{
		ServerEnv:  os.Getenv("SERVER_ENV"),
		DBName:     os.Getenv("DB_NAME"),
		DBHost:     os.Getenv("DB_HOST"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBPort:     port,
	}, nil
}
