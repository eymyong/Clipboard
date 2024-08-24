package main

import (
	"context"
	"fmt"
	"time"

	"github.com/eymyong/drop/model"
	"github.com/google/uuid"
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

	account := model.Account{
		Id:       uuid.NewString(),
		Username: "yyy",
		Password: "123",
	}

	data := make(map[string]string)
	//user = data[user]
	data["yyy"] = account.Id
	keyAccount := model.KeyAccount{
		Data: data,
	}

	fmt.Println("mapTest", keyAccount)

	//---------------------------------------

	// err := rd.HSet(ctx, "test:2", "id", "2").Err()
	// if err != nil {
	// 	panic(err)
	// }

	// err = rd.HSet(ctx, "test:2", "text", "two").Err()
	// if err != nil {
	// 	panic(err)
	// }

	// data1, err := rd.HGetAll(ctx, "test:2").Result()
	// if err != nil {
	// 	fmt.Printf("err: %w", err)
	// }
	// user := "2"
	// pass := "tw0"

	// for k, v := range data1 {
	// 	switch k {
	// 	case "id":
	// 		if v != user {
	// 			fmt.Println("not found username")
	// 			return
	// 		}
	// 	case "text":
	// 		if v != pass {
	// 			fmt.Println("password is incorrect")
	// 			return
	// 		}
	// 	}
	// }

	// fmt.Println("sucess")

	//-----------------------------------

	// d, err := rd.HGet(ctx, "test:2", "oo").Result()
	// if err != nil {
	// 	fmt.Printf("err: %w", err)
	// }

	// fmt.Println(d)

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

	// ctx := context.Background()
	// err := rd.Exists(ctx, "kkk").Err()
	// fmt.Println(err)

	// i, err := rd.Exists(ctx, "kkk").Result()
	// fmt.Println("i", i, "err", err)

}
