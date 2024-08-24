package repo

import (
	"context"

	"github.com/eymyong/drop/model"
)

type RepositoryClipboard interface {
	Create(ctx context.Context, clip model.Clipboard) error
	GetAll(ctx context.Context) ([]model.Clipboard, error)
	GetById(ctx context.Context, id string) (model.Clipboard, error)
	Update(ctx context.Context, id string, newdata string) error
	Delete(ctx context.Context, id string) error
}

type RepositoryUser interface {
	Create(ctx context.Context, user model.User) (model.User, error)
	GetPassword(ctx context.Context, username string) ([]byte, error)
}
