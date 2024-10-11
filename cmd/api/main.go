package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/eymyong/drop/cmd/api/config"
	"github.com/eymyong/drop/cmd/api/handler/auth"
	"github.com/eymyong/drop/cmd/api/handler/handlerclipboard"
	"github.com/eymyong/drop/cmd/api/handler/handleruser"
	"github.com/eymyong/drop/cmd/api/service"
	"github.com/eymyong/drop/repo"
	"github.com/eymyong/drop/repo/dbclipboard"
	"github.com/eymyong/drop/repo/redisclipboard"
	"github.com/eymyong/drop/repo/redisuser"
	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func main() {
	fileName := "config.json"
	conf := config.ReadJson(fileName)

	fmt.Println("conf:", conf)

	rd := repo.NewRedis(conf.RedisAddr, conf.RedisUsername, conf.RedisPassword, conf.RedisDb)

	confDB := config.DataSourceName(conf.DbHost, conf.DbPort, conf.DbUser, conf.DbName)
	fmt.Println("confD", confDB)

	db, err := sqlx.Connect("pgx", "host=167.179.66.149 port=5469 user=postgres dbname=yongdb")
	if err != nil {
		panic(err)
	}

	// db, err := repo.NewDb("pgx", "host=167.179.66.149 port=5469 user=postgres dbname=yongdb")
	// if err != nil {
	// 	panic(err)
	// }

	repoDB := dbclipboard.New(db)
	repoClip := redisclipboard.New(rd)
	repoUser := redisuser.New(rd)

	servicePassword := service.NewServicePassword(conf.SecretAES)
	authenticator := auth.New(conf.SecretJWT)

	// handlerClipDB := handlerclipboard.NewClipboard(repoDB)
	handlerClip := handlerclipboard.NewClipboard(repoClip, repoDB)
	handlerUser := handleruser.NewUser(repoUser, servicePassword, authenticator)

	r := newServer(handlerUser, handlerClip, authenticator)

	err = http.ListenAndServe(":8000", r)
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

	handlerclipboard.RegisterRoutesClipboardAPI(routerClip, handlerClip)
	handleruser.RegisterUserAPI(routerUsers, handlerUser)

	return r
}
