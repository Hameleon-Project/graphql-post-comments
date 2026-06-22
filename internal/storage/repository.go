package storage

import (
	"context"
	"graphql-post-comments/internal/model"
)

type PostRepository interface {
	CreatePost(ctx context.Context, post *model.Post) error
	UpdatePost(ctx context.Context, post *model.Post) error
	GetByID(ctx context.Context, id string) (*model.Post, error)
	GetAll(ctx context.Context) ([]*model.Post, error)
}

type CommentRepository interface {
	CreateComment(ctx context.Context, comment *model.Comment) error
	GetCommentByID(ctx context.Context, id string) (*model.Comment, error)
	GetByPostID(ctx context.Context, postID string, limit, offset int) ([]*model.Comment, error)
	GetCommentsByPostIDs(ctx context.Context, postIDs []string) (map[string][]*model.Comment, error)
}
