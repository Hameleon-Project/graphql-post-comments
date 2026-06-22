package service_test

import (
	"testing"
	"time"

	"graphql-post-comments/internal/model"
	"graphql-post-comments/internal/service"
)

func TestCommentEventBus_PublishSubscribe(t *testing.T) {
	bus := service.NewCommentEventBus()
	ch := bus.Subscribe("post1")

	comment := &model.Comment{ID: "c1", PostID: "post1", Content: "hello"}
	bus.Publish("post1", comment)

	select {
	case got := <-ch:
		if got.ID != "c1" {
			t.Fatalf("expected c1, got %s", got.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for comment")
	}

	bus.Unsubscribe("post1", ch)
}

func TestCommentEventBus_UnrelatedPost(t *testing.T) {
	bus := service.NewCommentEventBus()
	ch := bus.Subscribe("post1")

	bus.Publish("post2", &model.Comment{ID: "c2", PostID: "post2", Content: "other"})

	select {
	case <-ch:
		t.Fatal("should not receive comment for other post")
	case <-time.After(100 * time.Millisecond):
	}

	bus.Unsubscribe("post1", ch)
}
