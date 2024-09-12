package redisuser

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"

	"github.com/eymyong/drop/model"
	"github.com/eymyong/drop/repo"
)

const (
	keyLogins = "clipboard-logins"
	JwtKey    = "key"
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

func New(addr string, db int, pass string) repo.RepositoryUser {
	rd := redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       db,
		Password: pass,
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

	password, err := r.rd.HGet(ctx, key, "password").Result()
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

	//=============
	oldUsername, err := r.rd.HGet(ctx, key, "username").Result()
	if err != nil {
		return fmt.Errorf("hget old username redis err: %w", err)
	}

	password, err := r.rd.HGet(ctx, key, "password").Result()
	if err != nil {
		return fmt.Errorf("hget password redis err: %w", err)
	}

	err = r.rd.HDel(ctx, keyLogins, oldUsername).Err()
	if err != nil {
		return fmt.Errorf("hget old username redis err: %w", err)
	}

	err = r.rd.HSet(ctx, keyLogins, newUsername, password).Err()
	if err != nil {
		return fmt.Errorf("hset keylogin redis err: %w", err)
	}
	//=============

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

func NewJwt(userid, username string, key []byte) (token string, err error) {
	// TODO: investigate if Local() is actually needed
	exp := time.Now().Add(24 * time.Hour).Local()
	_ = exp
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Id:     userid,
		Issuer: username,
		// ExpiresAt: exp.Unix(),
	})
	// Generate JWT token from claims
	token, err = claims.SignedString(key)
	if err != nil {
		return token, errors.Wrapf(err, "failed to validate with key %s", key)
	}
	return token, nil
}

func VerifyJwt(tokenStr string, key []byte) (jwt.Claims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse JWT token %s", tokenStr)
	}

	return token.Claims, nil
}

func (r *RepoRedisUser) GetByUsername(ctx context.Context, username string) (string, error) {
	keys, err := r.rd.Keys(ctx, "users:*").Result()
	if err != nil {
		return "", fmt.Errorf("keys redis err: %w", err)
	}

	allId := []string{}
	for _, v := range keys {
		users, err := r.rd.HGetAll(ctx, v).Result()
		if err != nil {
			return "", fmt.Errorf("hgetall redis err: %w", err)
		}

		// map["id":ldjfogl,"username":yong,"password":"3333"]
		for k, v := range users {
			if k == "id" {
				allId = append(allId, v)
			}
		}

	}

	for _, id := range allId {
		keyId := keyUsers(id)
		uName, err := r.rd.HGet(ctx, keyId, "username").Result()
		if err != nil {
			return "", fmt.Errorf("hget username redis err: %w", err)
		}

		if uName == username {
			return id, nil
		}

		fmt.Println("uName:", uName)
		fmt.Println("username:", username)
	}

	return "", fmt.Errorf("not found username: %s", username)

}

func NewJwtTest(userid, username string, key []byte) (token string, err error) {
	// TODO: investigate if Local() is actually needed
	exp := time.Now().Add(24 * time.Hour).Local()
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Id:        userid,
		Issuer:    username,
		ExpiresAt: exp.Unix(),
	})
	// Generate JWT token from claims
	token, err = claims.SignedString(key)
	if err != nil {
		return token, errors.Wrapf(err, "failed to validate with key %s", key)
	}
	return token, nil
}
