package service

import (
	"context"
	"fmt"
	"time"

	"github.com/eymyong/drop/repo"
)

type User interface {
	Logout(ctx context.Context, token string, tokenExp int) error
}

type UserImpl struct {
	repoCache repo.RepositoryCaching
}

func NewServiceUser(repoCache repo.RepositoryCaching) *UserImpl {
	return &UserImpl{
		repoCache: repoCache,
	}
}

func (us *UserImpl) Logout(ctx context.Context, token string, tokenExp int) error {
	key := token
	now := time.Now()
	exp := time.Unix(int64(tokenExp), 0).Add(2 * time.Minute)
	dur := exp.Sub(now)
	ttl := dur.Seconds()

	err := us.repoCache.Create(ctx, key, now.String(), time.Duration(ttl))
	if err != nil {
		return fmt.Errorf("create cache err: %w", err)
	}

	return nil
}

// func f(a, b time.Time) int64 {
// 	au := a.Unix()
// 	bu := b.Unix()

// 	return bu - au
// }
