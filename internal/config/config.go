package config

import (
	"os"
)

type Config struct {
	StorageType string
	Port        string
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

	return &Config{
		StorageType: storage,
		Port:        port,
	}
}
