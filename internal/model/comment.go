package model

import (
	"errors"
	"time"
)

var (
	ErrCommentsDisabled = errors.New("comments are disabled for this post")
	ErrCommentTooLong   = errors.New("comment text exceeds 2000 characters")
)

type Comment struct {
	ID        string    `json:"id"`
	PostID    string    `json:"post_id"`
	ParentID  *string   `json:"parent_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
