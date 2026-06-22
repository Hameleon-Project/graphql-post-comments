package handler

import (
	"context"
	"net/http"

	"github.com/graph-gophers/dataloader/v7"

	"graphql-post-comments/internal/model"
	"graphql-post-comments/internal/service"
)

type contextKey string

const loadersKey contextKey = "dataloaders"

type Loaders struct {
	CommentsByPostID *dataloader.Loader[string, []*model.Comment]
}

func NewLoaders(commentService *service.CommentService) *Loaders {
	commentsLoader := dataloader.NewBatchedLoader(func(ctx context.Context, keys []string) []*dataloader.Result[[]*model.Comment] {
		commentsMap, err := commentService.GetCommentsByPostIDs(ctx, keys)
		results := make([]*dataloader.Result[[]*model.Comment], len(keys))
		for i, key := range keys {
			if err != nil {
				results[i] = &dataloader.Result[[]*model.Comment]{Error: err}
				continue
			}
			comments := commentsMap[key]
			if comments == nil {
				comments = []*model.Comment{}
			}
			results[i] = &dataloader.Result[[]*model.Comment]{Data: comments}
		}
		return results
	})

	return &Loaders{CommentsByPostID: commentsLoader}
}

func DataLoaderMiddleware(commentService *service.CommentService, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loaders := NewLoaders(commentService)
		ctx := context.WithValue(r.Context(), loadersKey, loaders)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetLoaders(ctx context.Context) *Loaders {
	loaders, _ := ctx.Value(loadersKey).(*Loaders)
	return loaders
}
