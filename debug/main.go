package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/eymyong/drop/model"
	"github.com/pkg/errors"
)

/*
redis-cli -h 167.179.66.149  -a ooBae5ciZiCohf9L

	Addr: "167.179.66.149:6379"
*/

func main() {

	clip := model.Clipboard{
		Id:     "1",
		UserId: "u1",
		Text:   "one",
	}
	if clip.Id != "" {
		fmt.Println("Test:", clip.Id)
	}
	fmt.Println("clip:", clip)

	clip2 := model.Clipboard{}
	if clip2.Id == "" && clip2.Text == "" && clip2.UserId == "" {
		slog.Error("failed to create in cache", "clipboard_id", clip.Id)
	}
	fmt.Println("clip2:", clip2)

	var a bool
	a = true
	if !a {
		fmt.Println("aF:", a)
	}

	if a {
		fmt.Println("aT:", a)
	}
	fmt.Println(a)

	// os.Getenv()
	//=================================================================================

	// rd := redis.NewClient(&redis.Options{
	// 	Addr: "127.0.0.1:6379",
	// 	// Password: "ooBae5ciZiCohf9L",
	// 	DB: 0,
	// })

	// ctx := context.Background()

	// err := rd.Set(ctx, "yong", "1234", 0).Err()
	// if err != nil {
	// 	fmt.Println("set err: %w", err)
	// }

	// data, err := rd.Get(ctx, "yong").Result()
	// if err != nil {
	// 	fmt.Errorf("get err: %w", err)
	// }

	// e, err := rd.Exists(ctx, "pak").Result()
	// if err != nil {
	// 	fmt.Errorf("exists err: %w", err)
	// }

	// fmt.Println("data:", data)
	// fmt.Println("e:", e)

	//================================================================================================
	// data, err := rd.Get(ctx, "yong").Result()
	// if err != nil {
	// 	fmt.Println("get err: %w", err)
	// }

	// fmt.Println("data\n", data)

	// err = rd.HSet(ctx, "test1", "id", "1", "name", "yong").Err()
	// if err != nil {
	// 	fmt.Println("hset err: %w", err)
	// }

	// data2, err := rd.HGetAll(ctx, "test1").Result()
	// if err != nil {
	// 	fmt.Println("hgetall err: %w", err)
	// }

	// fmt.Println("data2\n", data2)

	//==================================

	// id := "1"
	// username := "yong"
	// key := "keyTest"
	// k := []byte(key)

	// token, err := newJwt(id, username, k)
	// if err != nil {
	// 	fmt.Println("newJwt err: %w", err)
	// }
	// fmt.Println("token: ", token)

	// v, err := verifyJwt(token, k)
	// if err != nil {
	// 	fmt.Println("verifyJwt err: %w", err)
	// }
	// fmt.Println("v: ", v)

	// fmt.Println("==========")
	// tokenTest := "kdofj"
	// _ = tokenTest
	// kk := []byte("kkkk")
	// _ = kk

	// v2, err := verifyJwt(tokenTest, k)
	// if err != nil {
	// 	fmt.Println("verifyJwt err: %w", err)
	// }
	// fmt.Println("v2: ", v2)

	// s := v.Valid()
	// fmt.Println(s)

	//================================

	// secretJwt := "clipboard-jwt-secret"
	// expectedExp := time.Now().Add(24 * time.Hour).Local()
	// expectedId := "123"
	// expectedIss := "test"

	// claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
	// 	Id:        expectedId,
	// 	Issuer:    expectedIss,
	// 	ExpiresAt: expectedExp.Unix(),
	// })

	// token, err := claims.SignedString(secretJwt)
	// if err != nil {
	// 	fmt.Errorf("unexpect err:%w", err)
	// 	return
	// }

	// fmt.Println("token:", token)

	// if token == "" {
	// 	fmt.Errorf("expect token but got: %s", token)
	// 	return
	// }

	// actualClaoms, err := auth.ExtractClaims(token, []byte(secretJwt))
	// if err != nil {
	// 	fmt.Errorf("unexpect err: %w", err)
	// 	return
	// }

	// fmt.Println("actualClaoms", actualClaoms)

	// c, ok := actualClaoms.(jwt.MapClaims)
	// if !ok {
	// 	fmt.Errorf("unexpect ok: %v", ok)
	// }
	// _ = c

	//==================================================

}

func newJwt(userid, username string, key []byte) (token string, err error) {
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

func verifyJwt(tokenStr string, key []byte) (jwt.Claims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse JWT token %s", tokenStr)
	}

	return token.Claims, nil
}
