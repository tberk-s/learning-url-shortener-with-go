package main

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/tberk-s/learning-url-shortener-with-go/src/webserver"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	err := webserver.New(
		webserver.WithServerEnv(os.Getenv("SERVER_ENV")),
		webserver.WithDBName(os.Getenv("DB_NAME")),
		webserver.WithDBHost(os.Getenv("DB_HOST")),
		webserver.WithDBUser(os.Getenv("DB_USER")),
		webserver.WithDBPassword(os.Getenv("DB_PASSWORD")),
		webserver.WithDBPort(func() int {
			port, _ := strconv.Atoi(os.Getenv("DB_PORT"))
			return port
		}()),
	)
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
