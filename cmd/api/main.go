package main

import (
	"log"
	"net/http"

	"github.com/eymyong/drop/cmd/api/config"
	"github.com/eymyong/drop/cmd/api/handler/auth"
	"github.com/eymyong/drop/cmd/api/handler/handlerclipboard"
	"github.com/eymyong/drop/cmd/api/handler/handleruser"
	"github.com/eymyong/drop/cmd/api/service"
	"github.com/eymyong/drop/repo"
	"github.com/eymyong/drop/repo/redisclipboard"
	"github.com/eymyong/drop/repo/redisuser"
	"github.com/gorilla/mux"
)

func main() {
	fileName := "config.json"
	conf := config.ReadJson(fileName)

	rd := repo.NewRedis(conf.RedisAddr, conf.RedisUsername, conf.RedisPassword, conf.RedisDb)
	repoClip := redisclipboard.New(rd)
	repoUser := redisuser.New(rd)

	servicePassword := service.NewServicePassword(conf.SecretAES)
	authenticator := auth.New(conf.SecretJWT)

	handlerClip := handlerclipboard.NewClipboard(repoClip)
	handlerUser := handleruser.NewUser(repoUser, servicePassword, authenticator)

	r := newServer(handlerUser, handlerClip, authenticator)

	err := http.ListenAndServe(":8000", r)
	if err != nil {
		log.Println("server error:", err)
	}
}

func newServer(handlerUser *handleruser.HandlerUser, handlerClip *handlerclipboard.HandlerClipboard, auth *auth.AuthenticatorJWT) http.Handler {
	r := mux.NewRouter() // Main router

	routerClip := r.PathPrefix("/clipboards").Subrouter()
	routerUsers := r.PathPrefix("/users").Subrouter()

	r.HandleFunc("/register", handlerUser.Register).Methods(http.MethodPost)
	r.HandleFunc("/login", handlerUser.Login).Methods(http.MethodPost)

	routerClip.Use(auth.AuthMiddlewareHeader)
	routerUsers.Use(auth.AuthMiddlewareHeader)

	handlerclipboard.RegisterRoutesClipboardAPI(routerUsers, handlerClip)
	handleruser.RegisterUserAPI(routerClip, handlerUser)

	return r
}
