package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/eymyong/drop/repo"
	"github.com/redis/go-redis/v9"
)

type RepoCache struct {
	rd *redis.Client
}

func New(rd *redis.Client) repo.RepositoryCaching {
	return &RepoCache{rd: rd}
}

func (repo *RepoCache) Create(ctx context.Context, key string, value string, exp time.Duration) error {
	err := repo.rd.Set(ctx, key, value, exp).Err()
	if err != nil {
		return fmt.Errorf("redis set err: %w", err)
	}

	return nil
}

func (repo *RepoCache) BlacklistToken(ctx context.Context, tokenStr string) (int64, error) {

	blacklist, err := repo.rd.Exists(ctx, tokenStr).Result()
	if err != nil {
		return 0, fmt.Errorf("exists err: %w", err)
	}

	return blacklist, nil

}

// func (repo *RepoCache) BlacklistToken2(ctx context.Context, tokenStr string) (int64, error) {
// 	return repo.rd.Exists(ctx, tokenStr).Result()

// }
