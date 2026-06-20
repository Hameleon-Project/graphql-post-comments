package storage

import (
	"context"
	"errors"
	"sync"
	"time"

	"graphql-post-comments/internal/model"
)

type MemoryStorage struct {
	mu       sync.RWMutex
	posts    map[string]*model.Post
	comments map[string][]*model.Comment
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		posts:    make(map[string]*model.Post),
		comments: make(map[string][]*model.Comment),
	}
}

func (s *MemoryStorage) CreatePost(ctx context.Context, post *model.Post) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.posts[post.ID]; exists {
		return errors.New("post already exists")
	}
	s.posts[post.ID] = post
	return nil
}

func (s *MemoryStorage) GetByID(ctx context.Context, id string) (*model.Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	post, exists := s.posts[id]
	if !exists {
		return nil, errors.New("post not found")
	}
	return post, nil
}

func (s *MemoryStorage) GetAll(ctx context.Context) ([]*model.Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*model.Post, 0, len(s.posts))
	for _, post := range s.posts {
		result = append(result, post)
	}
	return result, nil
}

func (s *MemoryStorage) CreateComment(ctx context.Context, comment *model.Comment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	post, exists := s.posts[comment.PostID]
	if !exists {
		return errors.New("post not found")
	}
	if post.CommentsHidden {
		return model.ErrCommentsDisabled
	}

	if len(comment.Content) > 2000 {
		return model.ErrCommentTooLong
	}

	comment.CreatedAt = time.Now()
	s.comments[comment.PostID] = append(s.comments[comment.PostID], comment)
	return nil
}

func (s *MemoryStorage) GetByPostID(ctx context.Context, postID string, limit, offset int) ([]*model.Comment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	allComments, exists := s.comments[postID]
	if !exists {
		return []*model.Comment{}, nil
	}

	total := len(allComments)
	if offset > total {
		return []*model.Comment{}, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return allComments[offset:end], nil
}
