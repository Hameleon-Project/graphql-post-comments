package storage

import (
	"context"
	"errors"
	"sort"
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

func (s *MemoryStorage) UpdatePost(ctx context.Context, post *model.Post) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.posts[post.ID]; !exists {
		return model.ErrPostNotFound
	}
	s.posts[post.ID] = post
	return nil
}

func (s *MemoryStorage) GetByID(ctx context.Context, id string) (*model.Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	post, exists := s.posts[id]
	if !exists {
		return nil, model.ErrPostNotFound
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
		return model.ErrPostNotFound
	}
	if post.CommentsHidden {
		return model.ErrCommentsDisabled
	}

	if len(comment.Content) > model.MaxCommentLength {
		return model.ErrCommentTooLong
	}

	if comment.ParentID != nil {
		if err := s.validateParentLocked(comment.PostID, *comment.ParentID); err != nil {
			return err
		}
	}

	comment.CreatedAt = time.Now()
	s.comments[comment.PostID] = append(s.comments[comment.PostID], comment)
	return nil
}

func (s *MemoryStorage) GetCommentByID(ctx context.Context, id string) (*model.Comment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, comments := range s.comments {
		for _, c := range comments {
			if c.ID == id {
				return c, nil
			}
		}
	}
	return nil, model.ErrCommentNotFound
}

func (s *MemoryStorage) GetByPostID(ctx context.Context, postID string, limit, offset int) ([]*model.Comment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return paginateCommentTree(s.comments[postID], limit, offset), nil
}

func (s *MemoryStorage) GetCommentsByPostIDs(ctx context.Context, postIDs []string) (map[string][]*model.Comment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string][]*model.Comment, len(postIDs))
	for _, id := range postIDs {
		if comments, exists := s.comments[id]; exists {
			result[id] = comments
		} else {
			result[id] = []*model.Comment{}
		}
	}
	return result, nil
}

func (s *MemoryStorage) validateParentLocked(postID, parentID string) error {
	for _, c := range s.comments[postID] {
		if c.ID == parentID {
			return nil
		}
	}
	return model.ErrParentNotFound
}

func paginateCommentTree(allComments []*model.Comment, limit, offset int) []*model.Comment {
	if len(allComments) == 0 {
		return []*model.Comment{}
	}

	children := make(map[string][]*model.Comment)
	var roots []*model.Comment
	for _, c := range allComments {
		if c.ParentID == nil {
			roots = append(roots, c)
		} else {
			children[*c.ParentID] = append(children[*c.ParentID], c)
		}
	}

	sort.Slice(roots, func(i, j int) bool {
		return roots[i].CreatedAt.After(roots[j].CreatedAt)
	})

	if offset >= len(roots) {
		return []*model.Comment{}
	}

	end := len(roots)
	if limit > 0 {
		end = offset + limit
		if end > len(roots) {
			end = len(roots)
		}
	}
	paginatedRoots := roots[offset:end]

	result := make([]*model.Comment, 0)
	var collect func(*model.Comment)
	collect = func(c *model.Comment) {
		result = append(result, c)
		for _, child := range children[c.ID] {
			collect(child)
		}
	}
	for _, root := range paginatedRoots {
		collect(root)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})

	return result
}
