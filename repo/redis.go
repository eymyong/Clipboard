package repo

import (
	"github.com/redis/go-redis/v9"
)

func NewRedis(addr, username, password string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Username: username,
		Password: password,
		DB:       db,
	})
}
