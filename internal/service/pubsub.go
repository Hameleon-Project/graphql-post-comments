package service

import (
	"graphql-post-comments/internal/model"
	"sync"
)

type CommentEventBus struct {
	mu   sync.RWMutex
	subs map[string][]chan *model.Comment
}

func NewCommentEventBus() *CommentEventBus {
	return &CommentEventBus{
		subs: make(map[string][]chan *model.Comment),
	}
}

func (b *CommentEventBus) Subscribe(postID string) chan *model.Comment {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan *model.Comment, 1)
	b.subs[postID] = append(b.subs[postID], ch)
	return ch
}

func (b *CommentEventBus) Unsubscribe(postID string, ch chan *model.Comment) {
	b.mu.Lock()
	defer b.mu.Unlock()

	channels := b.subs[postID]
	for i, c := range channels {
		if c == ch {
			close(c)
			b.subs[postID] = append(channels[:i], channels[i+1:]...)
			break
		}
	}
	if len(b.subs[postID]) == 0 {
		delete(b.subs, postID)
	}
}

func (b *CommentEventBus) Publish(postID string, comment *model.Comment) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, ch := range b.subs[postID] {
		ch <- comment
	}
}
