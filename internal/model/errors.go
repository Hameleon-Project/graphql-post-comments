package model

import "errors"

const MaxCommentLength = 2000

var (
	ErrCommentsDisabled = errors.New("comments are disabled for this post")
	ErrCommentTooLong   = errors.New("comment exceeds maximum length of 2000 characters")
	ErrPostNotFound     = errors.New("post not found")
	ErrCommentNotFound  = errors.New("comment not found")
	ErrParentNotFound   = errors.New("parent comment not found or belongs to another post")
)
