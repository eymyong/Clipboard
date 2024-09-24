package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/eymyong/drop/cmd/api/handler"
	"github.com/eymyong/drop/cmd/api/handler/auth"
	"github.com/eymyong/drop/cmd/api/handler/handlerclipboard"
	"github.com/eymyong/drop/cmd/api/handler/handleruser"
	"github.com/eymyong/drop/cmd/api/service"
	"github.com/eymyong/drop/repo"
	"github.com/eymyong/drop/repo/redisclipboard"
	"github.com/eymyong/drop/repo/redisuser"
)

func main() {
	conf := envConfig()
	fmt.Println("conf", conf)

	rd := repo.NewRedis(conf.redisAddr, conf.redisUsername, conf.redisPassword, conf.redisDb)
	repoClip := redisclipboard.New(rd)
	repoUser := redisuser.New(rd)

	servicePassword := service.NewServicePassword(conf.secretAES)
	authenticator := auth.New(conf.secretJWT)

	handlerClip := handlerclipboard.NewClipboard(repoClip)
	handlerUser := handleruser.NewUser(repoUser, servicePassword, authenticator)
	// handlerUser := handleruser.NewUser(repoUser, servicePassword, authenticator)

	r := handler.NewServer(handlerUser, handlerClip, authenticator)

	err := http.ListenAndServe(":8000", r)
	if err != nil {
		log.Println("server error:", err)
	}
}
