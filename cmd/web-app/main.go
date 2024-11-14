package main

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/tberk-s/learning-url-shortener-with-go/src/apiserver"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	err := apiserver.New(
		apiserver.WithServerEnv(os.Getenv("SERVER_ENV")),
		apiserver.WithDBName(os.Getenv("DB_NAME")),
		apiserver.WithDBHost(os.Getenv("DB_HOST")),
		apiserver.WithDBUser(os.Getenv("DB_USER")),
		apiserver.WithDBPassword(os.Getenv("DB_PASSWORD")),
		apiserver.WithDBPort(func() int {
			port, _ := strconv.Atoi(os.Getenv("DB_PORT"))
			return port
		}()),
	)
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
