package caching

import "github.com/redis/go-redis/v9"

func NewRedisCaching(addr, username, password string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Username: username,
		Password: password,
		DB:       db,
	})
}
