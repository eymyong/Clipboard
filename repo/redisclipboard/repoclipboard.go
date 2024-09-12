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

func New(rd *redis.Client) repo.RepositoryClipboard {
	return &RepoRedis{rd: rd}
}

func (r *RepoRedis) Create(ctx context.Context, clip model.Clipboard) error {
	err := r.rd.HSet(ctx, keyRedisClipboard(clip.Id),
		"id", clip.Id,
		"text", clip.Text,
		"user_id", clip.UserId,
	).Err()
	if err != nil {
		return fmt.Errorf("hset redis err: %w", err)
	}
	return nil
}

func (r *RepoRedis) GetAll(ctx context.Context) ([]model.Clipboard, error) {
	keys, err := r.rd.Keys(ctx, "clipboard:*").Result()
	if err != nil {
		return []model.Clipboard{}, fmt.Errorf("keys redis err: %w", err)
	}

	clipboards := make([]model.Clipboard, len(keys))
	for i := range keys {
		data, err := r.rd.HGetAll(ctx, keys[i]).Result()
		if err != nil {
			return []model.Clipboard{}, fmt.Errorf("hgetall redis err: %w", err)
		}

		clip := model.Clipboard{
			Id:     data["id"],
			UserId: data["user_id"],
			Text:   data["text"],
		}

		clipboards[i] = clip
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

	return model.Clipboard{
		Id:     data["id"],
		UserId: data["user_id"],
		Text:   data["text"],
	}, nil
}

func (r *RepoRedis) Update(ctx context.Context, id string, newData string) error {
	key := keyRedisClipboard(id)
	c, err := r.rd.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("redis exists err: %w", err)
	}

	if c != 1 {
		return fmt.Errorf("unexpected length of redis keys %s: %d", key, c)
	}

	err = r.rd.HSet(ctx, key, "text", newData).Err()
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
