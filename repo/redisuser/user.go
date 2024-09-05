package redisuser

import (
	"context"
	"fmt"
	"strings"

	"github.com/eymyong/drop/model"
	"github.com/eymyong/drop/repo"
	"github.com/google/uuid"
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
	key := keyUsers(user.Username)
	count, err := r.rd.Exists(ctx, key).Result()
	if err != nil {
		return model.User{}, err
	}
	if count != 0 {
		return model.User{}, fmt.Errorf("username '%s' already exists", user.Username)
	}

	/* ควรจะใช้แบบไหน ต่างกันยังไง
	// userData := map[string]string{
	// 	"username": user.Username,
	// }

	// userDataJson, err := json.Marshal(userData)
	// if err != nil {
	// 	return model.User{}, err
	// }

	// err = r.rd.Set(ctx, key, string(userDataJson), 0).Err()
	// if err != nil {
	// 	return model.User{}, errors.Wrapf(err, "failed to save user '%s", user.Username)
	// }
	*/
	err = r.rd.Set(ctx, key, user.Username, 0).Err()
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

func (r *RepoRedisUser) GetAll(ctx context.Context) ([]model.User, error) {
	keys, err := r.rd.Keys(ctx, "users:*").Result()
	if err != nil {
		return []model.User{}, fmt.Errorf("keys redis err: %w", err)
	}

	users := []model.User{}
	for _, v := range keys {
		userName, err := r.rd.Get(ctx, v).Result()
		if err != nil {
			return []model.User{}, fmt.Errorf("get redis err")
		}

		passwordByte, err := r.GetPassword(ctx, keyToName(v))
		if err != nil {
			return []model.User{}, fmt.Errorf("get passsword err: %w", err)
		}

		user := model.User{
			Id:       uuid.NewString(),
			Username: userName,
			Password: string(passwordByte),
		}

		users = append(users, user)

	}

	return users, nil
}

func (r *RepoRedisUser) GetById(ctx context.Context, id string) (model.User, error) {
	// r.rd.

	panic(nil)
}

func (r *RepoRedisUser) UpdateUser(ctx context.Context, id string, newdata string) error {

	panic(nil)
}

func (r *RepoRedisUser) UpdatePassword(ctx context.Context, id string, newPassword string) error {

	panic(nil)
}

func (r *RepoRedisUser) Delete(ctx context.Context, id string) error {

	panic(nil)
}
