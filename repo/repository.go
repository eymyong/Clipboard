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
	Register(ctx context.Context, user string, age int, pass string) (model.KeyAccount, error)
	Login(ctx context.Context, user string, pass string) error
	//CreateUser(ctx context.Context, user model.User) error
	GetAllUser(ctx context.Context) ([]model.Account, error)
	GetByIdUser(ctx context.Context, id string) (model.Account, error)
	UpdateUser(ctx context.Context, id string, newdata string) error
	DeleteUser(ctx context.Context, id string) error
}
