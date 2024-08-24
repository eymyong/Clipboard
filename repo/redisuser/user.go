package redisuser

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/eymyong/drop/model"
	"github.com/eymyong/drop/repo"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

const (
	keyLogins = "clipboard-logins"
)

type RepoRedisUser struct {
	rd *redis.Client
}

func keyUsers(username string) string {
	return "users:" + username
}

func New(addr string, db int) repo.RepositoryUser {
	rd := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})

	return &RepoRedisUser{rd: rd}
}

func (r *RepoRedisUser) Create(ctx context.Context, user model.User) (model.User, error) {
	key := keyUsers(user.Username)
	count, err := r.rd.Exists(ctx, key).Result()
	if err != nil {
		return model.User{}, err
	}
	if count != 0 {
		return model.User{}, fmt.Errorf("username '%s' already exists", user.Username)
	}

	userData := map[string]string{
		"username": user.Username,
	}

	userDataJson, err := json.Marshal(userData)
	if err != nil {
		return model.User{}, err
	}

	err = r.rd.Set(ctx, key, string(userDataJson), 0).Err()
	if err != nil {
		return model.User{}, errors.Wrapf(err, "failed to save user '%s", user.Username)
	}

	err = r.rd.HSet(ctx, keyLogins, user.Username, user.Password).Err()
	if err != nil {
		return model.User{}, errors.Wrapf(err, "failed to save logins '%s", user.Username)
	}

	return user, nil
}

func (r *RepoRedisUser) GetPassword(ctx context.Context, username string) ([]byte, error) {
	pass, err := r.rd.HGet(ctx, keyLogins, username).Result()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get password for username '%s'", username)
	}

	return []byte(pass), nil
}
