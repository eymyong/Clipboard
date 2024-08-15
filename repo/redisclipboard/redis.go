package redisclipboard

import (
	"context"
	"fmt"

	"github.com/eymyong/drop/model"
	"github.com/eymyong/drop/repo"
	"github.com/redis/go-redis/v9"
)

type RepoRedis struct {
	rd *redis.Client
}

func keyRedisClipboard(id string) string {
	return "clipboard:" + id
}

func New(addr string) repo.Repository {
	rd := redis.NewClient(&redis.Options{
		Addr: addr,
		// DB:   db,
	})

	return &RepoRedis{rd: rd}
}

func (r *RepoRedis) Create(ctx context.Context, clip model.Clipboard) error {
	err := r.rd.HSet(ctx, keyRedisClipboard(clip.Id), "id", clip.Id, "text", clip.Text).Err()
	if err != nil {
		return fmt.Errorf("hset redis err: %w", err)
	}
	return nil
}

func (r *RepoRedis) GetAll(ctx context.Context) ([]model.Clipboard, error) {
	keyClipboards, err := r.rd.Keys(ctx, "*").Result()
	if err != nil {
		return []model.Clipboard{}, fmt.Errorf("keys redis err: %w", err)
	}

	clipboards := []model.Clipboard{}

	for _, v := range keyClipboards {
		data, err := r.rd.HGetAll(ctx, keyRedisClipboard(v)).Result()
		if err != nil {
			return []model.Clipboard{}, fmt.Errorf("hgetall redis err: %w", err)
		}

		clipboard := model.Clipboard{}
		for k, v := range data {
			switch k {
			case "id":
				clipboard.Id = v
			case "text":
				clipboard.Text = v
			default:
			}
		}
		clipboards = append(clipboards, clipboard)
	}

	return clipboards, nil
}

func (r *RepoRedis) GetById(ctx context.Context, id string) (model.Clipboard, error) {
	data, err := r.rd.HGetAll(ctx, keyRedisClipboard(id)).Result()
	if err != nil {
		return model.Clipboard{}, fmt.Errorf("hgetall redis err: %w", err)
	}

	clipboard := model.Clipboard{}
	for k, v := range data {
		switch k {
		case "id":
			clipboard.Id = v
		case "text":
			clipboard.Text = v
		default:
		}
	}

	return clipboard, nil
}
