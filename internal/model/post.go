package model

type Post struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Content        string `json:"content"`
	CommentsHidden bool   `json:"comments_hidden"`
}
