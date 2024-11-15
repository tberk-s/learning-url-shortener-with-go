package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	ServerEnv  string
	DBName     string
	DBHost     string
	DBUser     string
	DBPassword string
	DBPort     int
}

func LoadConfig() *Config {
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		log.Fatalf("Invalid database port: %v", err)
	}
	return &Config{
		ServerEnv:  os.Getenv("SERVER_ENV"),
		DBName:     os.Getenv("DB_NAME"),
		DBHost:     os.Getenv("DB_HOST"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBPort:     port,
	}
}
