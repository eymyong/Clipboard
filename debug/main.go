package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func main() {
	// repo := repo.RepositoryUser
	// ctx := context.Background()
	// rd := redisuser.New("127.0.0.1:6379", 3)

	// u := model.User{
	// 	Name: "yong",
	// 	Age:  20,
	// }
	// _ = u

	// // a, err := rd.Register(ctx, u)
	// // if err != nil {
	// // 	panic(err)
	// // }
	// // fmt.Println(a)

	// err := rd.Login(ctx, "", "yong", "1234")
	// if err != nil {
	// 	panic(err)
	// }
	// ------------------------
	rd := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	ctx := context.Background()

	users := "user: "
	type test struct {
		Id       string `json:"id"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	_ = users
	u := test{
		Id:       "1",
		Username: "one",
		Password: "1111",
	}
	_ = u

	username := "yong"
	_ = username

	err := rd.HSet(ctx, "test:1", "id", "1").Err()
	if err != nil {
		fmt.Println("hset redis err: %w", err)
	}

	value, err := rd.HGet(ctx, "test:1", "id").Result()
	if err != nil {
		fmt.Println("gset redis err: %w", err)
	}

	fmt.Println("value: ", value)

	value2, err := rd.HGet(ctx, "test:2", "id").Result()
	if err != nil {
		fmt.Println("gset redis err: %w", err)
	}

	fmt.Println("value2: ", value2)

	// keys, err := rd.Keys(ctx, "*").Result()
	// if err != nil {
	// 	panic(err)
	// }

	// m, err := rd.HGetAll(ctx, redisuser.KeyLogins).Result()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(m)

	// _, ok := m[username]
	// if ok {
	// 	fmt.Println("alalready exists")
	// 	return
	// }

	// fmt.Println("not found username")
}
