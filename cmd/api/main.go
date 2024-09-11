package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/eymyong/drop/cmd/api/handler/handlerclipboard"
	"github.com/eymyong/drop/cmd/api/handler/handleruser"
	"github.com/eymyong/drop/repo/redisclipboard"
	"github.com/eymyong/drop/repo/redisuser"
)

func envRedisDb() int {
	dbEnv, ok := os.LookupEnv("REDIS_DB")
	if !ok || dbEnv == "" {
		return 0
	}

	db, err := strconv.Atoi(dbEnv)
	if err != nil {
		return 0
	}

	return db
}

func envRedisAddr() string {
	const defaultAddr = "127.0.0.1:6379"

	addr, ok := os.LookupEnv("REDIS_ADDR")
	if !ok || addr == "" {
		return defaultAddr
	}

	return addr
}

func main() {
	redisAddr := envRedisAddr()
	redisDb := envRedisDb()

	repoClip := redisclipboard.New(redisAddr, redisDb)
	repoUser := redisuser.New(redisAddr, redisDb)

	hClip := handlerclipboard.NewClipboard(repoClip)
	hUser := handleruser.NewUser(repoUser)

	r := mux.NewRouter()

	r.HandleFunc("/clipboards/create", hClip.CreateClip).Methods(http.MethodPost)
	r.HandleFunc("/clipboards/getall", hClip.GetAllClips).Methods(http.MethodGet)
	r.HandleFunc("/clipboards/get/{clipboard-id}", hClip.GetClipById).Methods(http.MethodGet)
	r.HandleFunc("/clipboards/update/{clipboard-id}", hClip.UpdateClipById).Methods(http.MethodPatch)
	r.HandleFunc("/clipboards/delete/{clipboard-id}", hClip.DeleteClip).Methods(http.MethodDelete)

	r.HandleFunc("/users/register", hUser.Register).Methods(http.MethodPost)
	r.HandleFunc("/users/login", hUser.Login).Methods(http.MethodPost)
	r.HandleFunc("/users/get/{user-id}", hUser.GetUserById).Methods(http.MethodGet)
	r.HandleFunc("/users/update/username/{user-id}", hUser.UpdateUsername).Methods(http.MethodPatch)
	r.HandleFunc("/users/update/password/{user-id}", hUser.UpdatePassword).Methods(http.MethodPatch)
	r.HandleFunc("/users/delete/{user-id}", hUser.DeleteUser).Methods(http.MethodDelete)

	err := http.ListenAndServe(":8000", r)
	if err != nil {
		log.Println("server error:", err)
	}
}
