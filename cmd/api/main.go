package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/eymyong/drop/cmd/api/handler/handlerclipboard"
	"github.com/eymyong/drop/cmd/api/handler/handleruser"
	"github.com/eymyong/drop/cmd/api/handler/middlewares"
	"github.com/eymyong/drop/cmd/api/handler/middlewares/auth"
	"github.com/eymyong/drop/cmd/api/service"
	"github.com/eymyong/drop/repo"
	"github.com/eymyong/drop/repo/redisclipboard"
	"github.com/eymyong/drop/repo/redisuser"
)

func main() {
	conf := envConfig()
	log.Printf("config=%+v", conf)

	rd := repo.NewRedis(conf.redisAddr, conf.redisUsername, conf.redisPassword, conf.redisDb)
	repoClip := redisclipboard.New(rd)
	repoUser := redisuser.New(rd)

	servicePassword := service.NewServicePassword(conf.secretAES)
	authenticator := auth.New(conf.secretJWT)

	handlerClip := handlerclipboard.NewClipboard(repoClip)
	handlerUser := handleruser.NewUser(repoUser, servicePassword, authenticator)

	r := mux.NewRouter() // Main router
	r.Use(middlewares.NewLoggerV1(
		"clip-v2.3.6",
		[]string{
			"Authorization",
			"content-type",
		},
		true,
		true,
	))

	r.HandleFunc("/users/register", handlerUser.Register).Methods(http.MethodPost)
	r.HandleFunc("/users/login", handlerUser.Login).Methods(http.MethodPost)

	routerAccount := r.PathPrefix("/account").Subrouter()
	routerAccount.Use(authenticator.AuthMiddlewareHeader)

	routerAccount.HandleFunc("/get", handlerUser.GetUserById).Methods(http.MethodGet)
	routerAccount.HandleFunc("/update/username", handlerUser.UpdateUsername).Methods(http.MethodPatch)
	routerAccount.HandleFunc("/update/password", handlerUser.UpdatePassword).Methods(http.MethodPatch)
	routerAccount.HandleFunc("/delete", handlerUser.DeleteUser).Methods(http.MethodDelete)

	routerClipboards := r.PathPrefix("/clipboards").Subrouter()
	routerClipboards.Use(authenticator.AuthMiddlewareHeader)

	routerClipboards.HandleFunc("/create", handlerClip.CreateClip).Methods(http.MethodPost)
	routerClipboards.HandleFunc("/get-all", handlerClip.GetAllClips).Methods(http.MethodGet)
	routerClipboards.HandleFunc("/get/{clipboard-id}", handlerClip.GetClipById).Methods(http.MethodGet)
	routerClipboards.HandleFunc("/update/{clipboard-id}", handlerClip.UpdateClipById).Methods(http.MethodPatch)
	routerClipboards.HandleFunc("/delete/{clipboard-id}", handlerClip.DeleteClip).Methods(http.MethodDelete)

	log.Println("starting api")
	err := http.ListenAndServe(":8000", r)
	if err != nil {
		log.Println("server error:", err)
	}
}
