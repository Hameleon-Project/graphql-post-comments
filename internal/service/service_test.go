package service_test

import (
	"context"
	"strings"
	"testing"

	"graphql-post-comments/internal/model"
	"graphql-post-comments/internal/service"
	"graphql-post-comments/internal/storage"
)

func TestCommentService_CreateComment_Validation(t *testing.T) {
	ctx := context.Background()
	longContent := strings.Repeat("a", model.MaxCommentLength+1)
	missingParent := "missing"

	tests := []struct {
		name    string
		setup   func(t *testing.T, postSvc *service.PostService, commentSvc *service.CommentService) *model.Comment
		wantErr error
	}{
		{
			name: "comments disabled",
			setup: func(t *testing.T, postSvc *service.PostService, _ *service.CommentService) *model.Comment {
				t.Helper()
				if err := postSvc.CreatePost(ctx, &model.Post{
					ID: "p1", Title: "Title", Content: "Content", CommentsHidden: true,
				}); err != nil {
					t.Fatalf("CreatePost: %v", err)
				}
				return &model.Comment{ID: "c1", PostID: "p1", Content: "hello"}
			},
			wantErr: model.ErrCommentsDisabled,
		},
		{
			name: "comment too long",
			setup: func(t *testing.T, postSvc *service.PostService, _ *service.CommentService) *model.Comment {
				t.Helper()
				if err := postSvc.CreatePost(ctx, &model.Post{
					ID: "p1", Title: "Title", Content: "Content",
				}); err != nil {
					t.Fatalf("CreatePost: %v", err)
				}
				return &model.Comment{ID: "c1", PostID: "p1", Content: longContent}
			},
			wantErr: model.ErrCommentTooLong,
		},
		{
			name: "parent not found",
			setup: func(t *testing.T, postSvc *service.PostService, _ *service.CommentService) *model.Comment {
				t.Helper()
				if err := postSvc.CreatePost(ctx, &model.Post{
					ID: "p1", Title: "Title", Content: "Content",
				}); err != nil {
					t.Fatalf("CreatePost: %v", err)
				}
				return &model.Comment{ID: "c1", PostID: "p1", ParentID: &missingParent, Content: "hello"}
			},
			wantErr: model.ErrParentNotFound,
		},
		{
			name: "parent belongs to another post",
			setup: func(t *testing.T, postSvc *service.PostService, commentSvc *service.CommentService) *model.Comment {
				t.Helper()
				for _, p := range []*model.Post{
					{ID: "p1", Title: "Title", Content: "Content"},
					{ID: "p2", Title: "Other", Content: "Other"},
				} {
					if err := postSvc.CreatePost(ctx, p); err != nil {
						t.Fatalf("CreatePost: %v", err)
					}
				}
				if err := commentSvc.CreateComment(ctx, &model.Comment{ID: "c1", PostID: "p1", Content: "root"}); err != nil {
					t.Fatalf("CreateComment: %v", err)
				}
				parentID := "c1"
				return &model.Comment{ID: "c2", PostID: "p2", ParentID: &parentID, Content: "bad"}
			},
			wantErr: model.ErrParentNotFound,
		},
		{
			name: "post not found",
			setup: func(t *testing.T, _ *service.PostService, _ *service.CommentService) *model.Comment {
				t.Helper()
				return &model.Comment{ID: "c1", PostID: "missing", Content: "hello"}
			},
			wantErr: model.ErrPostNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemoryStorage()
			postSvc := service.NewPostService(store)
			commentSvc := service.NewCommentService(store, store)

			comment := tt.setup(t, postSvc, commentSvc)
			err := commentSvc.CreateComment(ctx, comment)
			if err != tt.wantErr {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestCommentService_CreateComment_Success(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStorage()
	postSvc := service.NewPostService(store)
	commentSvc := service.NewCommentService(store, store)

	if err := postSvc.CreatePost(ctx, &model.Post{ID: "p1", Title: "Title", Content: "Content"}); err != nil {
		t.Fatalf("CreatePost: %v", err)
	}

	if err := commentSvc.CreateComment(ctx, &model.Comment{ID: "c1", PostID: "p1", Content: "hello"}); err != nil {
		t.Fatalf("CreateComment: %v", err)
	}

	parentID := "c1"
	if err := commentSvc.CreateComment(ctx, &model.Comment{ID: "c2", PostID: "p1", ParentID: &parentID, Content: "reply"}); err != nil {
		t.Fatalf("CreateComment reply: %v", err)
	}

	comments, err := commentSvc.GetCommentsByPostID(ctx, "p1", 0, 0)
	if err != nil {
		t.Fatalf("GetCommentsByPostID: %v", err)
	}
	if len(comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(comments))
	}
}

func TestPostService_UpdatePost(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStorage()
	postSvc := service.NewPostService(store)

	post := &model.Post{ID: "p1", Title: "Title", Content: "Content", CommentsHidden: false}
	if err := postSvc.CreatePost(ctx, post); err != nil {
		t.Fatalf("CreatePost: %v", err)
	}

	post.CommentsHidden = true
	if err := postSvc.UpdatePost(ctx, post); err != nil {
		t.Fatalf("UpdatePost: %v", err)
	}

	got, err := postSvc.GetPostByID(ctx, "p1")
	if err != nil {
		t.Fatalf("GetPostByID: %v", err)
	}
	if !got.CommentsHidden {
		t.Fatal("expected comments hidden after toggle")
	}
}
