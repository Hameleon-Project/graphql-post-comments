package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"graphql-post-comments/internal/model"

	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(dataSourceName string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) CreatePost(ctx context.Context, post *model.Post) error {
	query := `INSERT INTO posts (id, title, content, comments_hidden) VALUES ($1, $2, $3, $4)`
	_, err := s.db.ExecContext(ctx, query, post.ID, post.Title, post.Content, post.CommentsHidden)
	return err
}

func (s *PostgresStorage) GetByID(ctx context.Context, id string) (*model.Post, error) {
	query := `SELECT id, title, content, comments_hidden FROM posts WHERE id = $1`
	post := &model.Post{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(&post.ID, &post.Title, &post.Content, &post.CommentsHidden)
	if err == sql.ErrNoRows {
		return nil, errors.New("post not found")
	}
	return post, err
}

func (s *PostgresStorage) GetAll(ctx context.Context) ([]*model.Post, error) {
	query := `SELECT id, title, content, comments_hidden FROM posts`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*model.Post
	for rows.Next() {
		p := &model.Post{}
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.CommentsHidden); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func (s *PostgresStorage) CreateComment(ctx context.Context, comment *model.Comment) error {

	var commentsHidden bool
	err := s.db.QueryRowContext(ctx, "SELECT comments_hidden FROM posts WHERE id = $1", comment.PostID).Scan(&commentsHidden)
	if err == sql.ErrNoRows {
		return errors.New("post not found")
	}
	if commentsHidden {
		return model.ErrCommentsDisabled
	}

	if len(comment.Content) > 2000 {
		return model.ErrCommentTooLong
	}

	comment.CreatedAt = time.Now()

	query := `INSERT INTO comments (id, post_id, parent_id, content, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err = s.db.ExecContext(ctx, query, comment.ID, comment.PostID, comment.ParentID, comment.Content, comment.CreatedAt)
	return err
}

func (s *PostgresStorage) GetByPostID(ctx context.Context, postID string, limit, offset int) ([]*model.Comment, error) {
	query := `
		WITH RECURSIVE comment_tree AS (
			-- Якорная часть: берем корневые комментарии для этого поста с пагинацией
			(
				SELECT id, post_id, parent_id, content, created_at
				FROM comments
				WHERE post_id = $1 AND parent_id IS NULL
				ORDER BY created_at DESC
				LIMIT $2 OFFSET $3
			)
			UNION ALL
			-- Рекурсивная часть: цепляем детей для выбранных комментов
			SELECT c.id, c.post_id, c.parent_id, c.content, c.created_at
			FROM comments c
			INNER JOIN comment_tree ct ON c.parent_id = ct.id
		)
		SELECT id, post_id, parent_id, content, created_at FROM comment_tree ORDER BY created_at ASC;
	`

	rows, err := s.db.QueryContext(ctx, query, postID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*model.Comment
	for rows.Next() {
		c := &model.Comment{}
		var parentID sql.NullString
		if err := rows.Scan(&c.ID, &c.PostID, &parentID, &c.Content, &c.CreatedAt); err != nil {
			return nil, err
		}
		if parentID.Valid {
			c.ParentID = &parentID.String
		} else {
			c.ParentID = nil
		}
		comments = append(comments, c)
	}
	return comments, nil
}
