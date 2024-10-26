package service

import (
	"context"
	"fmt"
	"time"

	"github.com/eymyong/drop/repo"
)

type User interface {
	Logout(ctx context.Context, token string, tokenExp int) error
	Blacklisted(ctx context.Context, token string) (bool, error)
}

type UserImpl struct {
	repoBlacklist repo.RepositoryBlacklist
}

func NewServiceUser(repoCache repo.RepositoryBlacklist) *UserImpl {
	return &UserImpl{
		repoBlacklist: repoCache,
	}
}

func (us *UserImpl) Logout(ctx context.Context, token string, tokenExp int) error {
	fmt.Println("logout token", token)
	key := token
	now := time.Now()
	exp := time.Unix(int64(tokenExp), 0).Add(2 * time.Minute)
	dur := exp.Sub(now)

	fmt.Println("dur", dur.String())

	err := us.repoBlacklist.Create(ctx, key, now.String(), dur)
	if err != nil {
		return fmt.Errorf("create cache err: %w", err)
	}

	return nil
}

func (us *UserImpl) Blacklisted(ctx context.Context, token string) (bool, error) {
	return us.repoBlacklist.IsBlacklisted(ctx, token)
}
