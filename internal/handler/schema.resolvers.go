package handler

import (
	"context"
	"graphql-post-comments/graph"
	"graphql-post-comments/internal/model"

	"github.com/google/uuid"
)

func (r *mutationResolver) CreatePost(ctx context.Context, title string, content string, commentsHidden bool) (*model.Post, error) {
	post := &model.Post{
		ID:             uuid.New().String(),
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
	post, err := r.Resolver.PostService.GetPostByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	if post.CommentsHidden {
		return nil, model.ErrCommentsDisabled
	}

	if len(content) > model.MaxCommentLength {
		return nil, model.ErrCommentTooLong
	}

	comment := &model.Comment{
		ID:       uuid.New().String(),
		PostID:   postID,
		ParentID: parentID,
		Content:  content,
	}

	if err := r.Resolver.CommentService.CreateComment(ctx, comment); err != nil {
		return nil, err
	}

	if r.Resolver.CommentBus != nil {
		r.Resolver.CommentBus.Publish(postID, comment)
	}

	return comment, nil
}

func (r *mutationResolver) TogglePostComments(ctx context.Context, id string, hidden bool) (*model.Post, error) {
	post, err := r.Resolver.PostService.GetPostByID(ctx, id)
	if err != nil {
		return nil, err
	}

	post.CommentsHidden = hidden

	if err := r.Resolver.PostService.UpdatePost(ctx, post); err != nil {
		return nil, err
	}

	return post, nil
}

func (r *postResolver) Comments(ctx context.Context, obj *model.Post, limit *int, offset *int) ([]*model.Comment, error) {
	if limit == nil && offset == nil {
		loaders := GetLoaders(ctx)
		if loaders != nil && loaders.CommentsByPostID != nil {
			thunk := loaders.CommentsByPostID.Load(ctx, obj.ID)
			return thunk()
		}
	}

	l := 0
	o := 0
	if limit != nil {
		l = *limit
	}
	if offset != nil {
		o = *offset
	}

	return r.Resolver.CommentService.GetCommentsByPostID(ctx, obj.ID, l, o)
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
	ch := r.Resolver.CommentBus.Subscribe(postID)

	go func() {
		<-ctx.Done()
		r.Resolver.CommentBus.Unsubscribe(postID, ch)
	}()

	return ch, nil
}

func (r *Resolver) Mutation() graph.MutationResolver { return &mutationResolver{r} }

func (r *Resolver) Post() graph.PostResolver { return &postResolver{r} }

func (r *Resolver) Query() graph.QueryResolver { return &queryResolver{r} }

func (r *Resolver) Subscription() graph.SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type postResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
