package redisclipboard

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"

	"github.com/eymyong/drop/model"
	"github.com/eymyong/drop/repo"
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

func (r *RepoRedis) GetAllUserClipboards(ctx context.Context, userId string) ([]model.Clipboard, error) {
	clips, err := r.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var results []model.Clipboard
	for _, clip := range clips {
		if clip.UserId != userId {
			continue
		}

		results = append(results, clip)
	}

	return clips, nil
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

func (r *RepoRedis) GetUserClipboard(ctx context.Context, id string, userId string) (model.Clipboard, error) {
	clip, err := r.GetById(ctx, id)
	if err != nil {
		return model.Clipboard{}, err
	}

	if clip.UserId != userId {
		return model.Clipboard{}, errors.Wrapf(redis.Nil, "no such clipboard for userID '%s'", userId)
	}

	return clip, nil
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

	n, err := r.rd.HSet(ctx, key, "text", newData).Result()
	if err != nil {
		return fmt.Errorf("hset redis err: %w", err)
	}
	if n != 1 {
		return fmt.Errorf("updated key %d != 1", n)
	}

	return nil
}

func (r *RepoRedis) UpdateUserClipboard(ctx context.Context, id string, userId string, text string) error {
	key := keyRedisClipboard(id)
	clipUserId, err := r.rd.HGet(ctx, key, "user_id").Result()
	if err != nil {
		return err
	}

	if clipUserId != userId {
		return errors.Wrapf(redis.Nil, "no such clipboard for userID '%s'", userId)
	}

	n, err := r.rd.HSet(ctx, key, "text", text).Result()
	if err != nil {
		return errors.Wrap(err, "failed to hset field 'text'")
	}

	if n != 1 {
		return fmt.Errorf("updated key %d != 1", n)
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

func (r *RepoRedis) DeleteUserClipboard(ctx context.Context, id string, userId string) error {
	key := keyRedisClipboard(id)
	clipUserId, err := r.rd.HGet(ctx, key, "user_id").Result()
	if err != nil {
		return err
	}

	if clipUserId != userId {
		return errors.Wrapf(redis.Nil, "no such clipboard for userID '%s'", userId)
	}

	return r.Delete(ctx, id)
}
