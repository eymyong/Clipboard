package main

import (
	"fmt"
)

func main() {
	type car struct {
		id   string
		name string
	}

	t := []car{}
	fmt.Println("data t:", t)
	fmt.Println("length t:", len(t))
}

/*
	rd := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		// Password: "ooBae5ciZiCohf9L",
		DB: 0,
	})

	ctx := context.Background()

	err := rd.Set(ctx, "test1", "tttt", 3600).Err()
	if err != nil {
		panic(err)
	}

	t, err := rd.Get(ctx, "test1").Result()
	if err != nil {
		panic(err)
	}

	fmt.Println(t)

	v, err := rd.Del(ctx, "test1").Result()
	if err != nil {
		panic(err)
	}

	fmt.Println(v)
*/
