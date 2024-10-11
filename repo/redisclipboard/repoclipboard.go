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

func New(rd *redis.Client) repo.RepositoryClipboard {
	return &RepoRedis{rd: rd}
}

func keyRedisClipboard(id string) string {
	return "clipboard:" + id
}

func (r *RepoRedis) Create(ctx context.Context, clip model.Clipboard) error {
	err := r.rd.HSet(ctx, keyRedisClipboard(clip.Id), "id", clip.Id, "userid", clip.UserId, "text", clip.Text).Err()
	if err != nil {
		return fmt.Errorf("hset redis err: %w", err)
	}
	return nil
}

func (r *RepoRedis) GetAll(ctx context.Context) ([]model.Clipboard, error) {
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
			case "userid":
				clipboard.UserId = v
			case "text":
				clipboard.Text = v
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

	if len(data) == 0 {
		return model.Clipboard{}, fmt.Errorf("no data in redis")
	}

	clipboard := model.Clipboard{}
	for k, v := range data {
		switch k {
		case "id":
			clipboard.Id = v
		case "userid":
			clipboard.UserId = v
		case "text":
			clipboard.Text = v
		}
	}

	return clipboard, nil
}

func (r *RepoRedis) Update(ctx context.Context, id string, newdata string) error {
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

func (r *RepoRedis) Delete(ctx context.Context, id string) error {
	err := r.rd.Del(ctx, keyRedisClipboard(id)).Err()
	if err != nil {
		return fmt.Errorf("del redis err: %w", err)
	}

	return nil
}

func (r *RepoRedis) DeleteAll(ctx context.Context) error {
	keys, err := r.rd.Keys(ctx, "clipboard:*").Result()
	if err != nil {
		return fmt.Errorf("keys redis err: %w", err)
	}

	for _, v := range keys {
		err := r.rd.Del(ctx, v).Err()
		if err != nil {
			return fmt.Errorf("del redis err: %w", err)
		}
	}

	keysAfter, err := r.rd.Keys(ctx, "clipboard:*").Result()
	if err != nil {
		return fmt.Errorf("keys redis err: %w", err)
	}

	if len(keysAfter) != 0 {
		return fmt.Errorf("found clipboard")
	}

	return nil
}
