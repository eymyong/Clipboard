package main

import (
	"context"
	"time"

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

	err := rd.Set(ctx, "t", "t1", time.Hour).Err()
	if err != nil {
		panic(err)
	}
}
