package storage_test

import (
	"context"
	"testing"

	"graphql-post-comments/internal/model"
	"graphql-post-comments/internal/storage"
)

func TestMemoryStorage_PostCRUD(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStorage()

	post := &model.Post{ID: "p1", Title: "Title", Content: "Content", CommentsHidden: false}
	if err := store.CreatePost(ctx, post); err != nil {
		t.Fatalf("CreatePost: %v", err)
	}

	got, err := store.GetByID(ctx, "p1")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Title != "Title" {
		t.Fatalf("expected Title, got %q", got.Title)
	}

	post.CommentsHidden = true
	if err := store.UpdatePost(ctx, post); err != nil {
		t.Fatalf("UpdatePost: %v", err)
	}

	got, err = store.GetByID(ctx, "p1")
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if !got.CommentsHidden {
		t.Fatal("expected comments to be hidden")
	}

	all, err := store.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 post, got %d", len(all))
	}
}

func TestMemoryStorage_CommentsDisabled(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStorage()

	post := &model.Post{ID: "p1", Title: "Title", Content: "Content", CommentsHidden: true}
	if err := store.CreatePost(ctx, post); err != nil {
		t.Fatalf("CreatePost: %v", err)
	}

	err := store.CreateComment(ctx, &model.Comment{ID: "c1", PostID: "p1", Content: "hello"})
	if err != model.ErrCommentsDisabled {
		t.Fatalf("expected ErrCommentsDisabled, got %v", err)
	}
}

func TestMemoryStorage_CommentTooLong(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStorage()

	post := &model.Post{ID: "p1", Title: "Title", Content: "Content"}
	if err := store.CreatePost(ctx, post); err != nil {
		t.Fatalf("CreatePost: %v", err)
	}

	longContent := make([]byte, model.MaxCommentLength+1)
	for i := range longContent {
		longContent[i] = 'a'
	}

	err := store.CreateComment(ctx, &model.Comment{ID: "c1", PostID: "p1", Content: string(longContent)})
	if err != model.ErrCommentTooLong {
		t.Fatalf("expected ErrCommentTooLong, got %v", err)
	}
}

func TestMemoryStorage_HierarchicalPagination(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStorage()

	post := &model.Post{ID: "p1", Title: "Title", Content: "Content"}
	if err := store.CreatePost(ctx, post); err != nil {
		t.Fatalf("CreatePost: %v", err)
	}

	parent1 := "root1"
	parent2 := "root2"
	parent3 := "root3"

	for _, c := range []model.Comment{
		{ID: parent1, PostID: "p1", Content: "root 1"},
		{ID: parent2, PostID: "p1", Content: "root 2"},
		{ID: parent3, PostID: "p1", Content: "root 3"},
		{ID: "child3", PostID: "p1", ParentID: &parent3, Content: "reply to root3"},
	} {
		if err := store.CreateComment(ctx, &c); err != nil {
			t.Fatalf("CreateComment %s: %v", c.ID, err)
		}
	}

	page, err := store.GetByPostID(ctx, "p1", 1, 0)
	if err != nil {
		t.Fatalf("GetByPostID: %v", err)
	}
	if len(page) != 2 {
		t.Fatalf("expected 2 comments (1 root + 1 child), got %d", len(page))
	}

	ids := map[string]bool{}
	for _, c := range page {
		ids[c.ID] = true
	}
	if !ids[parent3] || !ids["child3"] {
		t.Fatalf("expected newest root3 and child3, got %v", ids)
	}
}

func TestMemoryStorage_InvalidParent(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStorage()

	post := &model.Post{ID: "p1", Title: "Title", Content: "Content"}
	if err := store.CreatePost(ctx, post); err != nil {
		t.Fatalf("CreatePost: %v", err)
	}

	parentID := "missing"
	err := store.CreateComment(ctx, &model.Comment{ID: "c1", PostID: "p1", ParentID: &parentID, Content: "hello"})
	if err != model.ErrParentNotFound {
		t.Fatalf("expected ErrParentNotFound, got %v", err)
	}
}

func TestMemoryStorage_GetCommentsByPostIDs(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStorage()

	for _, id := range []string{"p1", "p2"} {
		if err := store.CreatePost(ctx, &model.Post{ID: id, Title: id, Content: id}); err != nil {
			t.Fatalf("CreatePost: %v", err)
		}
	}

	if err := store.CreateComment(ctx, &model.Comment{ID: "c1", PostID: "p1", Content: "a"}); err != nil {
		t.Fatalf("CreateComment: %v", err)
	}

	result, err := store.GetCommentsByPostIDs(ctx, []string{"p1", "p2", "p3"})
	if err != nil {
		t.Fatalf("GetCommentsByPostIDs: %v", err)
	}
	if len(result["p1"]) != 1 {
		t.Fatalf("expected 1 comment for p1, got %d", len(result["p1"]))
	}
	if len(result["p2"]) != 0 {
		t.Fatalf("expected 0 comments for p2, got %d", len(result["p2"]))
	}
	if len(result["p3"]) != 0 {
		t.Fatalf("expected 0 comments for p3, got %d", len(result["p3"]))
	}
}
