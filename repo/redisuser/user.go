package redisuser

import (
	"context"
	"fmt"
	"time"

	"github.com/eymyong/drop/model"
	"github.com/eymyong/drop/repo"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RepoRedisUser struct {
	rd *redis.Client
}

func keyRedisAccount(id string) string {
	return "Account:" + id
}

func New(addr string, db int) repo.RepositoryUser {
	rd := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})

	return &RepoRedisUser{rd: rd}
}

func (r *RepoRedisUser) Register(ctx context.Context, user string, age int, pass string) (model.KeyAccount, error) {
	if age < 18 {
		return model.KeyAccount{}, fmt.Errorf("applicants under age")
	}

	account := model.Account{
		Id:       uuid.NewString(),
		Username: user,
		Password: pass,
	}
	err := r.rd.HSet(ctx, keyRedisAccount(account.Id), "Id", account.Id).Err()
	if err != nil {
		return model.KeyAccount{}, fmt.Errorf("hset redis err: %w", err)
	}

	err = r.rd.HSet(ctx, keyRedisAccount(account.Id), "Username", account.Username).Err()
	if err != nil {
		return model.KeyAccount{}, fmt.Errorf("hset redis err: %w", err)
	}

	err = r.rd.HSet(ctx, keyRedisAccount(account.Id), "Password", account.Password).Err()
	if err != nil {
		return model.KeyAccount{}, fmt.Errorf("hset redis err: %w", err)
	}

	data := make(map[string]string)
	data[user] = account.Id
	keyAccount := model.KeyAccount{
		Data: data,
	}

	err = r.rd.Set(ctx, account.Username, account.Id, time.Hour).Err()
	if err != nil {
		return model.KeyAccount{}, fmt.Errorf("hset redis err: %w", err)
	}

	return keyAccount, nil
}

// key: [Account:11111 Id:11111 Username:yong Password:0000]
func (r *RepoRedisUser) Login(ctx context.Context, user string, pass string) error {
	id, err := r.rd.Get(ctx, user).Result()
	if err != nil {
		return fmt.Errorf("get redis err: %w", err)
	}

	datas, err := r.rd.HGetAll(ctx, keyRedisAccount(id)).Result()
	if err != nil {
		return fmt.Errorf("hgetall redis err: %w", err)
	}

	for k, v := range datas {
		switch k {
		case "Username":
			if v != user {
				return fmt.Errorf("not found username")
			}
		case "Password":
			if v != pass {
				return fmt.Errorf("password is incorrect")
			}
		}
	}

	fmt.Println("login success")
	return nil

	// p, err := r.rd.HGet(ctx, keyRedisAccount(id), "Password").Result()
	// if err != nil {
	// 	return fmt.Errorf("hget redis err: %w", err)
	// }

	// if pass != p {
	// 	return fmt.Errorf("password is incorrect")
	// }

	// fmt.Println("success")

	// return nil

}

// func (r *RepoRedisUser) CreateUser(ctx context.Context, user model.User) error {
// 	r.rd.HSet(ctx,keyRedisAccount(user.Name))

// 	return nil
// }

func (r *RepoRedisUser) GetAllUser(ctx context.Context) ([]model.Account, error) {
	keys, err := r.rd.Keys(ctx, "Account:*").Result()
	if err != nil {
		return []model.Account{}, fmt.Errorf("keys redis err: %w", err)
	}

	accounts := []model.Account{}
	for _, v := range keys {
		data, err := r.rd.HGetAll(ctx, v).Result()
		if err != nil {
			return []model.Account{}, fmt.Errorf("hgetall redis err: %w", err)
		}

		account := model.Account{}
		for k, v := range data {
			switch k {
			case "id":
				account.Id = v
			case "Username":
				account.Username = v
			case "Password":
				account.Password = v
			}
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (r *RepoRedisUser) GetByIdUser(ctx context.Context, id string) (model.Account, error) {

	return model.Account{}, nil
}

func (r *RepoRedisUser) UpdateUser(ctx context.Context, id string, newdata string) error {

	return nil
}

func (r *RepoRedisUser) DeleteUser(ctx context.Context, id string) error {

	return nil
}
