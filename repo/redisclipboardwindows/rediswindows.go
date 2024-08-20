package redisclipboardwindows

import (
	"context"
	"fmt"

	"github.com/eymyong/drop/model"
	"github.com/eymyong/drop/repo"
	"github.com/redis/go-redis/v9"
)

type RepoRedisWindows struct {
	rd *redis.Client
}

func keyRedisClipboard(id string) string {
	return "clipboard:" + id
}

func New(addr string, db int) repo.Repository {
	rd := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})

	return &RepoRedisWindows{rd: rd}
}

func (r *RepoRedisWindows) Create(ctx context.Context, clip model.Clipboard) error {
	key := keyRedisClipboard(clip.Id)
	err := r.rd.HSet(ctx, key, "id", clip.Id).Err()
	if err != nil {
		return fmt.Errorf("hset redis field `id` err: %w", err)
	}

	err = r.rd.HSet(ctx, key, "text", clip.Text).Err()
	if err != nil {
		return fmt.Errorf("hset redis field `text` err: %w", err)
	}
	return nil
}

func (r *RepoRedisWindows) GetAll(ctx context.Context) ([]model.Clipboard, error) {
	keyClipboards, err := r.rd.Keys(ctx, "clipboard:*").Result()
	if err != nil {
		return []model.Clipboard{}, fmt.Errorf("keys redis err: %w", err)
	}

	clipboards := []model.Clipboard{}
	for _, v := range keyClipboards {
		data, err := r.rd.HGetAll(ctx, v).Result()
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
			}
		}
		clipboards = append(clipboards, clipboard)
	}

	return clipboards, nil
}

func (r *RepoRedisWindows) GetById(ctx context.Context, id string) (model.Clipboard, error) {
	data, err := r.rd.HGetAll(ctx, keyRedisClipboard(id)).Result()
	if err != nil {
		return model.Clipboard{}, fmt.Errorf("hgetall redis err: %w", err)
	}

	if len(data) == 0 {
		return model.Clipboard{}, fmt.Errorf("no data in redis")
	}

	clipboard := model.Clipboard{}
	for k, v := range data {
		switch k {
		case "id":
			clipboard.Id = v
		case "text":
			clipboard.Text = v
		}
	}

	return clipboard, nil
}

func (r *RepoRedisWindows) Update(ctx context.Context, id string, newdata string) error {
	key := keyRedisClipboard(id)
	c, err := r.rd.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("redis exists err: %w", err)
	}

	if c != 1 {
		return fmt.Errorf("unexpected length of redis keys %s: %d", key, c)
	}

	err = r.rd.HSet(ctx, key, "text", newdata).Err()
	if err != nil {
		return fmt.Errorf("hset redis err: %w", err)
	}

	return nil
}

func (r *RepoRedisWindows) Delete(ctx context.Context, id string) error {
	err := r.rd.Del(ctx, keyRedisClipboard(id)).Err()
	if err != nil {
		return fmt.Errorf("del redis err: %w", err)
	}

	return nil
}
