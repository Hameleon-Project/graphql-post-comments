package config

import (
	"os"
)

type Config struct {
	StorageType string
	Port        string
	DatabaseURL string
}

func Load() *Config {
	storage := os.Getenv("STORAGE_TYPE")
	if storage == "" {
		storage = "memory"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/comments_db?sslmode=disable"
	}

	return &Config{
		StorageType: storage,
		Port:        port,
		DatabaseURL: dbURL,
	}
}
