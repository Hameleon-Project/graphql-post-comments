package main

import (
	"fmt"
	"log"
	"net/http"

	gqlhandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"graphql-post-comments/graph"
	"graphql-post-comments/internal/config"
	"graphql-post-comments/internal/handler"
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
		if err := pgStorage.Migrate("migrations"); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		postRepo = pgStorage
		commentRepo = pgStorage
		log.Println("PostgreSQL storage initialized successfully")
	} else {
		log.Fatalf("Unknown storage type: %s", cfg.StorageType)
	}

	postService := service.NewPostService(postRepo)
	commentService := service.NewCommentService(commentRepo)
	commentBus := service.NewCommentEventBus()

	srv := gqlhandler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{
		Resolvers: &handler.Resolver{
			PostService:    postService,
			CommentService: commentService,
			CommentBus:     commentBus,
		},
	}))

	dataloaderHandler := handler.DataLoaderMiddleware(commentService, srv)

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", dataloaderHandler)

	log.Printf("Connect to http://localhost:%s/ for GraphQL playground", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
