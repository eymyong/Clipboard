package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/eymyong/drop/cmd/api/handler/auth"
	"github.com/eymyong/drop/cmd/api/handler/handlerclipboard"
	"github.com/eymyong/drop/cmd/api/handler/handleruser"
	"github.com/eymyong/drop/cmd/api/service"
	"github.com/eymyong/drop/repo"
	"github.com/eymyong/drop/repo/redisclipboard"
	"github.com/eymyong/drop/repo/redisuser"
)

func envRedisDb() int {
	const defaultDb = 0

	dbEnv, ok := os.LookupEnv("REDIS_DB")
	if !ok || dbEnv == "" {
		return defaultDb
	}

	db, err := strconv.Atoi(dbEnv)
	if err != nil {
		return defaultDb
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

func envRedisUsername() string {
	const defaultUsername = ""

	rdUsername, ok := os.LookupEnv("REDIS_USERNAME")
	if !ok || rdUsername == "" {
		return defaultUsername
	}

	return rdUsername
}

func envRedisPassword() string {
	const defaultPassword = ""

	rdPassword, ok := os.LookupEnv("REDIS_PASSWORD")
	if !ok || rdPassword == "" {
		return defaultPassword
	}

	return rdPassword
}

func envPasswordKeyAES() string {
	const defaultKey = "my-secret-foobarbaz200030004000x"

	k, ok := os.LookupEnv("CLIP_PASSWORD_KEY_AES")
	if !ok || len(k) < 32 {
		return defaultKey
	}

	return k
}

func envSecretJWT() string {
	const defaultSecret = "clipboard-jwt-secret"

	s, ok := os.LookupEnv("CLIP_SECRET_JWT")
	if !ok || s == "" {
		return defaultSecret
	}

	return s
}

func whoAmI(w http.ResponseWriter, r *http.Request) {
	id := r.Header.Get("id")

	if id == "" {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "missing header 'id'")
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "the user id is '%s'", id)
}

func main() {
	redisAddr := envRedisAddr()
	redisDb := envRedisDb()
	redisUsername := envRedisUsername()
	redisPassword := envRedisPassword()
	passwordKeyAES := envPasswordKeyAES()
	secretJWT := envSecretJWT()

	rd := repo.NewRedis(redisAddr, redisUsername, redisPassword, redisDb)
	repoClip := redisclipboard.New(rd)
	repoUser := redisuser.New(rd)

	servicePassword := service.NewServicePassword(passwordKeyAES)
	authenticator := auth.New(secretJWT)

	hClip := handlerclipboard.NewClipboard(repoClip)
	hUser := handleruser.NewUser(repoUser, servicePassword, authenticator)

	r := mux.NewRouter()               // Main router
	j := r.Path("/whoami").Subrouter() // Router for testing JWT

	j.Use(authenticator.AuthMiddlewareBody)
	j.HandleFunc("", whoAmI)

	r.HandleFunc("/clipboards/create", hClip.CreateClip).Methods(http.MethodPost)
	r.HandleFunc("/clipboards/get-all", hClip.GetAllClips).Methods(http.MethodGet)
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
