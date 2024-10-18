package redisuser

import (
	"context"
	"fmt"
	"log"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"

	"github.com/eymyong/drop/model"
	"github.com/eymyong/drop/repo"
)

const (
	keyRedisUsers = "clipboard-users" // Redis hash, field=userID, value=JSON data

	mapUsernamePassword = "clipboard-username-password" // Redis hash, field=username, value=password
	mapUsernameUserID   = "clipboard-username-ids"      // Redis hash, field=username, value=userID
)

type RepoRedisUser struct {
	rd *redis.Client
}

func userKey(id string) string {
	return keyRedisUsers + ":" + id
}

func New(rd *redis.Client) repo.RepositoryUser {
	return &RepoRedisUser{rd: rd}
}

func (r *RepoRedisUser) Create(ctx context.Context, user model.User) (model.User, error) {
	dup, err := r.DuplicateUserId(ctx, user.Id)
	if err != nil {
		return model.User{}, errors.Wrap(err, "failed to check for duplicate user id")
	}

	if dup {
		return model.User{}, errors.New("the new user id is already taken")
	}

	dup, err = r.DuplicateUsername(ctx, user.Username)
	if err != nil {
		return model.User{}, errors.Wrapf(err, "failed to check forr duplicate username '%s'", user.Username)
	}

	if dup {
		return model.User{}, fmt.Errorf("username '%s' is already taken", user.Username)
	}

	key := userKey(user.Id)
	err = r.rd.HSet(ctx, key, map[string]interface{}{
		"id":       user.Id,
		"username": user.Username,
		"password": user.Password,
	}).Err()

	if err != nil {
		return model.User{}, errors.Wrapf(err, "failed to register user '%s", user.Username)
	}

	err1 := r.rd.HSet(ctx, mapUsernamePassword, user.Username, user.Password).Err()
	err2 := r.rd.HSet(ctx, mapUsernameUserID, user.Username, user.Id).Err()

	if err1 != nil || err2 != nil {
		if err1 != nil {
			log.Printf("error when setting logins for user '%s': %s", user.Username, err1.Error())
		}

		if err2 != nil {
			log.Printf("error when setting userID mapping for user '%s': %s", user.Username, err2.Error())
		}

		_ = r.rd.HDel(ctx, mapUsernamePassword, user.Username).Err()
		_ = r.rd.HDel(ctx, mapUsernameUserID, user.Username).Err()

		return model.User{}, errors.Wrapf(err, "failed to create mappings for user '%s", user.Username)
	}

	return user, nil
}

func (r *RepoRedisUser) GetPassword(ctx context.Context, username string) ([]byte, error) {
	pass, err := r.rd.HGet(ctx, mapUsernamePassword, username).Result()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get password for username '%s'", username)
	}

	return []byte(pass), nil
}

func (r *RepoRedisUser) GetById(ctx context.Context, id string) (model.User, error) {
	key := userKey(id)
	username, err := r.rd.HGet(ctx, key, "username").Result()
	if err != nil {
		return model.User{}, fmt.Errorf("get redis err: %w", err)
	}

	password, err := r.rd.HGet(ctx, mapUsernamePassword, username).Result()
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

func (r *RepoRedisUser) GetUserId(ctx context.Context, username string) (string, error) {
	id, err := r.rd.HGet(ctx, mapUsernameUserID, username).Result()
	if err != nil {
		return "", err
	}

	return id, nil
}

func (r *RepoRedisUser) DuplicateUserId(ctx context.Context, id string) (bool, error) {
	key := userKey(id)
	count, err := r.rd.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *RepoRedisUser) DuplicateUsername(ctx context.Context, username string) (bool, error) {
	exists, err := r.rd.HExists(ctx, mapUsernamePassword, username).Result()
	if err != nil {
		return false, fmt.Errorf("hgetall redis err: %w", err)
	}

	return exists, nil
}

func (r *RepoRedisUser) UpdateUsername(ctx context.Context, id string, newUsername string) error {
	return errors.New("not implemented: updateUsername")
}

func (r *RepoRedisUser) UpdatePassword(ctx context.Context, id string, newPassword string) error {
	return errors.New("not implemented: updatePassword")
}

func (r *RepoRedisUser) Delete(ctx context.Context, id string) error {
	return errors.New("not implemented: deleteUser")
}
