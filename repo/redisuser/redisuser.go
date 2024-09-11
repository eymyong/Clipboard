package redisuser

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"

	"github.com/eymyong/drop/model"
	"github.com/eymyong/drop/repo"
)

const (
	keyLogins = "clipboard-logins"
)

type RepoRedisUser struct {
	rd *redis.Client
}

func keyUsers(id string) string {
	return "users:" + id
}

// "users:yong"
func keyToName(key string) string {
	word := strings.Split(key, ":")

	return word[1]
}

func New(addr string, db int) repo.RepositoryUser {
	rd := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})

	return &RepoRedisUser{rd: rd}
}

func (r *RepoRedisUser) Create(ctx context.Context, user model.User) (model.User, error) {
	dup, err := r.duplicateUserId(ctx, user.Id)
	if err != nil {
		return model.User{}, errors.Wrap(err, "failed to check for duplicate user id")
	}

	if dup {
		return model.User{}, errors.New("the new user id is already taken")
	}

	dup, err = r.duplicateUsername(ctx, user.Username)
	if err != nil {
		return model.User{}, errors.Wrapf(err, "failed to check forr duplicate username '%s'", user.Username)
	}

	if dup {
		return model.User{}, fmt.Errorf("username '%s' is already taken", user.Username)
	}

	key := keyUsers(user.Id)
	err = r.rd.HSet(ctx, key, map[string]interface{}{
		"id":       user.Id,
		"username": user.Username,
		"password": user.Password,
	}).Err()
	if err != nil {
		return model.User{}, errors.Wrapf(err, "failed to register user '%s", user.Username)
	}

	err = r.rd.HSet(ctx, keyLogins, user.Username, user.Password).Err()
	if err != nil {
		return model.User{}, errors.Wrapf(err, "failed to create logins for user '%s", user.Username)
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

func (r *RepoRedisUser) GetById(ctx context.Context, id string) (model.User, error) {
	key := keyUsers(id)
	username, err := r.rd.HGet(ctx, key, "username").Result()
	if err != nil {
		return model.User{}, fmt.Errorf("get redis err: %w", err)
	}

	password, err := r.rd.HGet(ctx, keyLogins, username).Result()
	if err != nil {
		return model.User{}, fmt.Errorf("hget redis in getbyid err: %w", err)
	}

	user := model.User{
		Id:       id,
		Username: username,
		Password: password,
	}

	return user, nil
}

func (r *RepoRedisUser) UpdateUsername(ctx context.Context, id string, newUsername string) error {
	key := keyUsers(id)

	dup, err := r.duplicateUsername(ctx, newUsername)
	if err != nil {
		return errors.Wrapf(err, "failed to check for duplicate new username '%s'", newUsername)
	}

	if dup {
		return fmt.Errorf("username %s is already taken", newUsername)
	}

	err = r.rd.HSet(ctx, key, "username", newUsername).Err()
	if err != nil {
		return fmt.Errorf("hset new username redis err: %w", err)
	}

	return nil
}

func (r *RepoRedisUser) UpdatePassword(ctx context.Context, id string, newPassword string) error {
	panic("not implemented")
}

func (r *RepoRedisUser) Delete(ctx context.Context, id string) error {
	username, err := r.rd.HGet(ctx, keyUsers(id), "username").Result()
	if err != nil {
		return fmt.Errorf("failed to get username: %w", err)
	}

	count, err := r.rd.Del(ctx, keyUsers(id)).Result()
	if err != nil {
		return fmt.Errorf("del redis err: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("0 user deleted for id '%s'", id)
	}

	count, err = r.rd.HDel(ctx, keyLogins, username).Result()
	if err != nil {
		return fmt.Errorf("del keylogins redis err: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("0 logins deleted for id '%s'", id)
	}

	return nil
}

func (r *RepoRedisUser) duplicateUserId(ctx context.Context, id string) (bool, error) {
	key := keyUsers(id)
	count, err := r.rd.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *RepoRedisUser) duplicateUsername(ctx context.Context, username string) (bool, error) {
	exists, err := r.rd.HExists(ctx, keyLogins, username).Result()
	if err != nil {
		return false, fmt.Errorf("hgetall redis err: %w", err)
	}

	return exists, nil
}
