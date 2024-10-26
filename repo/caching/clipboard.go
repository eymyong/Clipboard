package caching

import (
	"context"
	"log/slog"

	"github.com/eymyong/drop/model"
	"github.com/eymyong/drop/repo"
)

type repoCache interface {
	Sync(ctx context.Context, clip model.Clipboard) error
	repo.RepositoryClipboard
}

type RepoCachingClipboard struct {
	cache repoCache
	db    repo.RepositoryClipboard
}

func (r *RepoCachingClipboard) Create(ctx context.Context, clip model.Clipboard) error {
	err := r.db.Create(ctx, clip)
	if err != nil {
		return err
	}

	err = r.cache.Sync(ctx, clip)
	if err != nil {
		slog.Error("failed to create in cache", "clipboard_id", clip.Id)
	}

	return nil
}

func (r *RepoCachingClipboard) GetAll(ctx context.Context) ([]model.Clipboard, error) {
	return nil, nil
}

func (r *RepoCachingClipboard) GetById(ctx context.Context, id string) (model.Clipboard, error)
func (r *RepoCachingClipboard) Update(ctx context.Context, id string, newdata string) error
func (r *RepoCachingClipboard) Delete(ctx context.Context, id string) error
func (r *RepoCachingClipboard) DeleteAll(ctx context.Context) error
func (r *RepoCachingClipboard) GetAllUserClipboards(ctx context.Context, userId string) ([]model.Clipboard, error)
func (r *RepoCachingClipboard) GetUserClipboard(ctx context.Context, id string, userId string) (model.Clipboard, error)
func (r *RepoCachingClipboard) UpdateUserClipboard(ctx context.Context, id string, userId string, text string) error
func (r *RepoCachingClipboard) DeleteUserClipboard(ctx context.Context, id string, userId string) error
