package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func main() {

	rd := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	ctx := context.Background()
	err := rd.Exists(ctx, "kkk").Err()
	fmt.Println(err)

	i, err := rd.Exists(ctx, "kkk").Result()
	fmt.Println("i", i, "err", err)

	// // err := rd.HSet(ctx, "test:1", "id", "2", "data", "two", "status", "DONE").Err()
	// // if err != nil {
	// // 	panic(err)
	// // }

	// err := rd.HSet(ctx, "test:1", "Id", 1, "text", "one").Err()
	// if err != nil {
	// 	panic(err)
	// }

	// data, err := rd.HGetAll(ctx, "test:1").Result()
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println(data)
	// data := model.Clipboard{
	// 	Id:   "111",
	// 	Text: "one",
	// }

	// r := redisclipboard.New("127.0.0.1:6379")
	// ctx := context.Background()

	// err := r.Create(ctx, data)
	// if err != nil {
	// 	panic(err)
	// }

}
