package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"graphql-post-comments/internal/model"

	"github.com/lib/pq"
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

func (s *PostgresStorage) Migrate(migrationsDir string) error {
	path := filepath.Join(migrationsDir, "000001_init.up.sql")
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read migration: %w", err)
	}
	if _, err := s.db.Exec(string(content)); err != nil {
		return fmt.Errorf("apply migration: %w", err)
	}
	return nil
}

func (s *PostgresStorage) CreatePost(ctx context.Context, post *model.Post) error {
	query := `INSERT INTO posts (id, title, content, comments_hidden) VALUES ($1, $2, $3, $4)`
	_, err := s.db.ExecContext(ctx, query, post.ID, post.Title, post.Content, post.CommentsHidden)
	return err
}

func (s *PostgresStorage) UpdatePost(ctx context.Context, post *model.Post) error {
	query := `UPDATE posts SET title = $1, content = $2, comments_hidden = $3 WHERE id = $4`
	result, err := s.db.ExecContext(ctx, query, post.Title, post.Content, post.CommentsHidden, post.ID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return model.ErrPostNotFound
	}
	return nil
}

func (s *PostgresStorage) GetByID(ctx context.Context, id string) (*model.Post, error) {
	query := `SELECT id, title, content, comments_hidden FROM posts WHERE id = $1`
	post := &model.Post{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(&post.ID, &post.Title, &post.Content, &post.CommentsHidden)
	if err == sql.ErrNoRows {
		return nil, model.ErrPostNotFound
	}
	return post, err
}

func (s *PostgresStorage) GetAll(ctx context.Context) ([]*model.Post, error) {
	query := `SELECT id, title, content, comments_hidden FROM posts ORDER BY id`
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
	return posts, rows.Err()
}

func (s *PostgresStorage) CreateComment(ctx context.Context, comment *model.Comment) error {
	comment.CreatedAt = time.Now()

	query := `INSERT INTO comments (id, post_id, parent_id, content, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := s.db.ExecContext(ctx, query, comment.ID, comment.PostID, comment.ParentID, comment.Content, comment.CreatedAt)
	return err
}

func (s *PostgresStorage) GetCommentByID(ctx context.Context, id string) (*model.Comment, error) {
	query := `SELECT id, post_id, parent_id, content, created_at FROM comments WHERE id = $1`
	c := &model.Comment{}
	var parentID sql.NullString
	err := s.db.QueryRowContext(ctx, query, id).Scan(&c.ID, &c.PostID, &parentID, &c.Content, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, model.ErrCommentNotFound
	}
	if err != nil {
		return nil, err
	}
	if parentID.Valid {
		c.ParentID = &parentID.String
	}
	return c, nil
}

func (s *PostgresStorage) GetByPostID(ctx context.Context, postID string, limit, offset int) ([]*model.Comment, error) {
	var query string
	var args []any

	if limit > 0 {
		query = `
			WITH RECURSIVE comment_tree AS (
				(
					SELECT id, post_id, parent_id, content, created_at
					FROM comments
					WHERE post_id = $1 AND parent_id IS NULL
					ORDER BY created_at DESC
					LIMIT $2 OFFSET $3
				)
				UNION ALL
				SELECT c.id, c.post_id, c.parent_id, c.content, c.created_at
				FROM comments c
				INNER JOIN comment_tree ct ON c.parent_id = ct.id
			)
			SELECT id, post_id, parent_id, content, created_at FROM comment_tree ORDER BY created_at ASC`
		args = []any{postID, limit, offset}
	} else {
		query = `
			WITH RECURSIVE comment_tree AS (
				(
					SELECT id, post_id, parent_id, content, created_at
					FROM comments
					WHERE post_id = $1 AND parent_id IS NULL
					ORDER BY created_at DESC
				)
				UNION ALL
				SELECT c.id, c.post_id, c.parent_id, c.content, c.created_at
				FROM comments c
				INNER JOIN comment_tree ct ON c.parent_id = ct.id
			)
			SELECT id, post_id, parent_id, content, created_at FROM comment_tree ORDER BY created_at ASC`
		args = []any{postID}
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanComments(rows)
}

func (s *PostgresStorage) GetCommentsByPostIDs(ctx context.Context, postIDs []string) (map[string][]*model.Comment, error) {
	if len(postIDs) == 0 {
		return make(map[string][]*model.Comment), nil
	}

	query := `
		SELECT id, post_id, parent_id, content, created_at
		FROM comments
		WHERE post_id = ANY($1)
		ORDER BY created_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query, pq.Array(postIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]*model.Comment)
	for rows.Next() {
		c := &model.Comment{}
		var parentID sql.NullString
		if err := rows.Scan(&c.ID, &c.PostID, &parentID, &c.Content, &c.CreatedAt); err != nil {
			return nil, err
		}
		if parentID.Valid {
			c.ParentID = &parentID.String
		}
		result[c.PostID] = append(result[c.PostID], c)
	}
	return result, rows.Err()
}

func scanComments(rows *sql.Rows) ([]*model.Comment, error) {
	var comments []*model.Comment
	for rows.Next() {
		c := &model.Comment{}
		var parentID sql.NullString
		if err := rows.Scan(&c.ID, &c.PostID, &parentID, &c.Content, &c.CreatedAt); err != nil {
			return nil, err
		}
		if parentID.Valid {
			c.ParentID = &parentID.String
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}
