package main

import (
	"context"
	"encoding/json"
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

	userData := map[string]string{
		"username": "test1",
	}

	userDataJson, err := json.Marshal(userData)
	if err != nil {
		panic("marshal err")
	}

	// err := rd.Set(ctx, "t", "t1", time.Hour).Err()
	// if err != nil {
	// 	panic(err)
	// }

	// i, err := rd.Exists(ctx, "ttt").Result()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(i)
	err = rd.Set(ctx, "user:test1", string(userDataJson), 0).Err()
	if err != nil {
		panic("set err")
	}

	data, err := rd.Get(ctx, "user:test1").Result()
	fmt.Println(data)

	data2, err := rd.HGetAll(ctx, "t").Result()
	if err != nil {
		panic("getall err")
	}
	fmt.Println(data2)
}
