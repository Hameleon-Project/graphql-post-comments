package handler

import (
	"context"
	"errors"
	"fmt"
	"graphql-post-comments/graph"
	"graphql-post-comments/internal/model"
	"time"
)

func (r *mutationResolver) CreatePost(ctx context.Context, title string, content string, commentsHidden bool) (*model.Post, error) {
	id := fmt.Sprintf("%d", time.Now().UnixNano())

	post := &model.Post{
		ID:             id,
		Title:          title,
		Content:        content,
		CommentsHidden: commentsHidden,
	}

	if err := r.Resolver.PostService.CreatePost(ctx, post); err != nil {
		return nil, err
	}

	return post, nil
}

func (r *mutationResolver) CreateComment(ctx context.Context, postID string, parentID *string, content string) (*model.Comment, error) {
	if len(content) > 2000 {
		return nil, model.ErrCommentTooLong
	}

	id := fmt.Sprintf("%d", time.Now().UnixNano())

	comment := &model.Comment{
		ID:       id,
		PostID:   postID,
		ParentID: parentID,
		Content:  content,
	}

	if err := r.Resolver.CommentService.CreateComment(ctx, comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (r *mutationResolver) ToggleComments(ctx context.Context, postID string, hidden bool) (*model.Post, error) {
	post, err := r.Resolver.PostService.GetPostByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	post.CommentsHidden = hidden

	if err := r.Resolver.PostService.CreatePost(ctx, post); err != nil {
		return nil, err
	}

	return post, nil
}

func (r *postResolver) Comments(ctx context.Context, obj *model.Post, limit int, offset int) ([]*model.Comment, error) {
	return r.Resolver.CommentService.GetCommentsByPostID(ctx, obj.ID, limit, offset)
}

func (r *queryResolver) Posts(ctx context.Context) ([]*model.Post, error) {
	return r.Resolver.PostService.GetAllPosts(ctx)
}

func (r *queryResolver) Post(ctx context.Context, id string) (*model.Post, error) {
	post, err := r.Resolver.PostService.GetPostByID(ctx, id)
	if err != nil {
		return nil, nil
	}
	return post, nil
}

func (r *subscriptionResolver) CommentAdded(ctx context.Context, postID string) (<-chan *model.Comment, error) {
	return nil, errors.New("subscriptions are not implemented yet")
}

func (r *Resolver) Mutation() graph.MutationResolver { return &mutationResolver{r} }

func (r *Resolver) Post() graph.PostResolver { return &postResolver{r} }

func (r *Resolver) Query() graph.QueryResolver { return &queryResolver{r} }

func (r *Resolver) Subscription() graph.SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type postResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
