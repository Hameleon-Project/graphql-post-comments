package service_test

import (
	"context"
	"testing"

	"graphql-post-comments/internal/model"
	"graphql-post-comments/internal/service"
	"graphql-post-comments/internal/storage"
)

func TestCommentService_ValidateParent(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStorage()
	postSvc := service.NewPostService(store)
	commentSvc := service.NewCommentService(store)

	post := &model.Post{ID: "p1", Title: "Title", Content: "Content"}
	if err := postSvc.CreatePost(ctx, post); err != nil {
		t.Fatalf("CreatePost: %v", err)
	}

	otherPost := &model.Post{ID: "p2", Title: "Other", Content: "Other"}
	if err := postSvc.CreatePost(ctx, otherPost); err != nil {
		t.Fatalf("CreatePost p2: %v", err)
	}

	if err := commentSvc.CreateComment(ctx, &model.Comment{ID: "c1", PostID: "p1", Content: "root"}); err != nil {
		t.Fatalf("CreateComment: %v", err)
	}

	wrongPostParent := "c1"
	err := commentSvc.CreateComment(ctx, &model.Comment{ID: "c2", PostID: "p2", ParentID: &wrongPostParent, Content: "bad"})
	if err != model.ErrParentNotFound {
		t.Fatalf("expected ErrParentNotFound, got %v", err)
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
