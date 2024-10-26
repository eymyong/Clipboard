package blacklist

import (
	"context"
	"fmt"
	"time"

	"github.com/eymyong/drop/repo"
	"github.com/redis/go-redis/v9"
)

type RepoBlacklist struct {
	rd *redis.Client
}

func New(rd *redis.Client) repo.RepositoryBlacklist {
	return &RepoBlacklist{rd: rd}
}

func (repo *RepoBlacklist) Create(ctx context.Context, key string, value string, exp time.Duration) error {
	err := repo.rd.Set(ctx, key, value, exp).Err()
	if err != nil {
		return fmt.Errorf("redis set err: %w", err)
	}

	return nil
}

func (repo *RepoBlacklist) IsBlacklisted(ctx context.Context, tokenStr string) (bool, error) {
	blacklist, err := repo.rd.Exists(ctx, tokenStr).Result()
	if err != nil {
		return false, fmt.Errorf("exists err: %w", err)
	}

	return blacklist == 1, nil
}
