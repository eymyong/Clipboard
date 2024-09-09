package redisuser

import (
	"context"
	"fmt"
	"strings"

	"github.com/eymyong/drop/model"
	"github.com/eymyong/drop/repo"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

const (
	KeyLogins = "clipboard-logins"
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
	key := keyUsers(user.Id)
	count, err := r.rd.Exists(ctx, key).Result()
	if err != nil {
		return model.User{}, err
	}

	if count != 0 {
		return model.User{}, fmt.Errorf("id '%s' already exists", user.Id)
	}

	err = r.CheckDuplicateName(user.Username)
	if err != nil {
		return model.User{}, fmt.Errorf("create err: %w", err)
	}

	err = r.rd.HSet(ctx, key, "id", user.Id).Err()
	if err != nil {
		return model.User{}, errors.Wrapf(err, "failed to save user id '%s", user.Id)
	}

	err = r.rd.HSet(ctx, key, "username", user.Username).Err()
	if err != nil {
		return model.User{}, errors.Wrapf(err, "failed to save username '%s", user.Username)
	}

	err = r.rd.HSet(ctx, key, "password", user.Password).Err()
	if err != nil {
		return model.User{}, errors.Wrapf(err, "failed to save user password '%s", user.Password)
	}

	err = r.rd.HSet(ctx, KeyLogins, user.Username, user.Password).Err()
	if err != nil {
		return model.User{}, errors.Wrapf(err, "failed to save logins '%s", user.Username)
	}

	return user, nil
}

func (r *RepoRedisUser) GetPassword(ctx context.Context, username string) ([]byte, error) {
	pass, err := r.rd.HGet(ctx, KeyLogins, username).Result()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get password for username '%s'", username)
	}

	return []byte(pass), nil
}

func (r *RepoRedisUser) GetAll(ctx context.Context) ([]model.User, error) {
	keys, err := r.rd.Keys(ctx, "users:*").Result()
	if err != nil {
		return []model.User{}, fmt.Errorf("keys redis err: %w", err)
	}

	users := []model.User{}
	for _, v := range keys {

		id, err := r.rd.HGet(ctx, v, "id").Result()
		if err != nil {
			return []model.User{}, fmt.Errorf("hget id redis err: %w", err)
		}

		username, err := r.rd.HGet(ctx, v, "username").Result()
		if err != nil {
			return []model.User{}, fmt.Errorf("hget username redis err: %w", err)
		}

		user := model.User{
			Id:       id,
			Username: username,
			Password: "*****",
		}

		users = append(users, user)

	}

	return users, nil
}

func (r *RepoRedisUser) GetById(ctx context.Context, id string) (model.User, error) {
	key := keyUsers(id)
	username, err := r.rd.HGet(ctx, key, "username").Result()
	if err != nil {
		return model.User{}, fmt.Errorf("get redis err: %w", err)
	}

	password, err := r.rd.HGet(ctx, KeyLogins, username).Result()
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

/*
func (r *RepoRedisUser) CheckDuplicateName(username string) (bool, error) {
	ctx := context.Background()
	data, err := r.rd.HGetAll(ctx, KeyLogins).Result()
	if err != nil {
		return false, fmt.Errorf("hgetall redis err: %w")
	}

	_, ok := data[username]
	if ok {
		return false, fmt.Errorf("username already taken")
	}

	return true, nil
}
*/

func (r *RepoRedisUser) CheckDuplicateName(username string) error {
	ctx := context.Background()
	data, err := r.rd.HGetAll(ctx, KeyLogins).Result()
	if err != nil {
		return fmt.Errorf("hgetall redis err: %w", err)
	}

	_, ok := data[username]
	if ok {
		return fmt.Errorf("username already taken")
	}

	return nil
}

func (r *RepoRedisUser) UpdateUserName(ctx context.Context, id string, newUsername string) error {
	key := keyUsers(id)

	err := r.CheckDuplicateName(newUsername)
	if err != nil {
		return fmt.Errorf("duplicate: %w", err)
		// return fmt.Errorf("check username err: %w", err)
	}

	user, err := r.rd.HGetAll(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("hgetall redis err: %w", err)
	}

	for k, v := range user {
		switch k {
		case "username":
			err := r.rd.HDel(ctx, KeyLogins, v).Err()
			if err != nil {
				return fmt.Errorf("del redis err: %w", err)
			}
		case "password":
			err := r.rd.HSet(ctx, KeyLogins, newUsername, v).Err()
			if err != nil {
				return fmt.Errorf("hset redis err: %w", err)
			}
		}
	}

	err = r.rd.HSet(ctx, key, "username", newUsername).Err()
	if err != nil {
		return fmt.Errorf("hset new username redis err: %w", err)
	}

	return nil
}

func (r *RepoRedisUser) UpdatePassword(ctx context.Context, id string, newPassword string) error {
	key := keyUsers(id)

	username, err := r.rd.HGet(ctx, key, "username").Result()
	if err != nil {
		return fmt.Errorf("hget username redis err: %w", err)
	}

	if username == "" {
		return fmt.Errorf("not found username")
	}

	err = r.rd.HSet(ctx, key, "password", newPassword).Err()
	if err != nil {
		return fmt.Errorf("hset newpassword redis err: %w", err)
	}

	err = r.rd.HSet(ctx, KeyLogins, username, newPassword).Err()
	if err != nil {
		return fmt.Errorf("hset keylogins redis err: %w", err)
	}

	return nil
}

func (r *RepoRedisUser) Delete(ctx context.Context, id string) error {

	username, err := r.rd.HGet(ctx, keyUsers(id), "username").Result()
	if err != nil {
		return fmt.Errorf("hget redis err: %w", err)
	}

	err = r.rd.Del(ctx, keyUsers(id)).Err()
	if err != nil {
		return fmt.Errorf("del redis err: %w", err)
	}

	err = r.rd.HDel(ctx, KeyLogins, username).Err()
	if err != nil {
		return fmt.Errorf("del keylogins redis err: %w", err)
	}

	return nil
}

func (r *RepoRedisUser) DeleteAll(ctx context.Context) error {
	keys, err := r.rd.Keys(ctx, "*").Result()
	if err != nil {
		return fmt.Errorf("keys redis err: %w", err)
	}

	for _, v := range keys {
		err := r.rd.Del(ctx, v).Err()
		if err != nil {
			return fmt.Errorf("del redis err: %w", err)
		}
	}

	return nil
}
