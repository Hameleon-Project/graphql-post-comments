package handler

import "graphql-post-comments/internal/service"

type Resolver struct {
	PostService    *service.PostService
	CommentService *service.CommentService
	CommentBus     *service.CommentEventBus
}
