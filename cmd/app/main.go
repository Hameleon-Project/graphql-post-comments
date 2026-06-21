package main

import (
	"fmt"
	"log"

	"graphql-post-comments/internal/config"
	"graphql-post-comments/internal/service"
	"graphql-post-comments/internal/storage"
)

func main() {
	cfg := config.Load()

	fmt.Printf("Starting service with storage type: %s on port: %s\n", cfg.StorageType, cfg.Port)

	var postRepo storage.PostRepository
	var commentRepo storage.CommentRepository

	if cfg.StorageType == "memory" {
		memStorage := storage.NewMemoryStorage()
		postRepo = memStorage
		commentRepo = memStorage
		log.Println("In-memory storage initialized successfully")
	} else if cfg.StorageType == "postgres" {
		pgStorage, err := storage.NewPostgresStorage(cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("Failed to initialize postgres storage: %v", err)
		}
		postRepo = pgStorage
		commentRepo = pgStorage
		log.Println("PostgreSQL storage initialized successfully")
	} else {
		log.Fatalf("Unknown storage type: %s", cfg.StorageType)
	}

	postService := service.NewPostService(postRepo)
	commentService := service.NewCommentService(commentRepo)

	_ = postService
	_ = commentService

	log.Println("Services successfully initialized. Ready for transport layer.")
}
