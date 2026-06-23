package service

import (
	"context"

	"graphql-post-comments/internal/model"
	"graphql-post-comments/internal/storage"
)

type PostService struct {
	repo storage.PostRepository
}

func NewPostService(repo storage.PostRepository) *PostService {
	return &PostService{repo: repo}
}

func (s *PostService) CreatePost(ctx context.Context, post *model.Post) error {
	return s.repo.CreatePost(ctx, post)
}

func (s *PostService) UpdatePost(ctx context.Context, post *model.Post) error {
	return s.repo.UpdatePost(ctx, post)
}

func (s *PostService) GetPostByID(ctx context.Context, id string) (*model.Post, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *PostService) GetAllPosts(ctx context.Context) ([]*model.Post, error) {
	return s.repo.GetAll(ctx)
}

type CommentService struct {
	postRepo    storage.PostRepository
	commentRepo storage.CommentRepository
}

func NewCommentService(postRepo storage.PostRepository, commentRepo storage.CommentRepository) *CommentService {
	return &CommentService{postRepo: postRepo, commentRepo: commentRepo}
}

func (s *CommentService) CreateComment(ctx context.Context, comment *model.Comment) error {
	post, err := s.postRepo.GetByID(ctx, comment.PostID)
	if err != nil {
		return err
	}
	if post.CommentsHidden {
		return model.ErrCommentsDisabled
	}
	if len(comment.Content) > model.MaxCommentLength {
		return model.ErrCommentTooLong
	}
	if comment.ParentID != nil {
		parent, err := s.commentRepo.GetCommentByID(ctx, *comment.ParentID)
		if err != nil {
			return model.ErrParentNotFound
		}
		if parent.PostID != comment.PostID {
			return model.ErrParentNotFound
		}
	}
	return s.commentRepo.CreateComment(ctx, comment)
}

func (s *CommentService) GetCommentsByPostID(ctx context.Context, postID string, limit, offset int) ([]*model.Comment, error) {
	return s.commentRepo.GetByPostID(ctx, postID, limit, offset)
}

func (s *CommentService) GetCommentsByPostIDs(ctx context.Context, postIDs []string) (map[string][]*model.Comment, error) {
	return s.commentRepo.GetCommentsByPostIDs(ctx, postIDs)
}
