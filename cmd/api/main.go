package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/eymyong/drop/cmd/api/handler/handlerclipboard"
	"github.com/eymyong/drop/cmd/api/handler/handleruser"
	"github.com/eymyong/drop/cmd/api/service"
	"github.com/eymyong/drop/repo/redisclipboard"
	"github.com/eymyong/drop/repo/redisuser"
)

/*
redis-cli -h 167.179.66.149  -a ooBae5ciZiCohf9L
*/

func envRedisAddr() string {
	const defaultAddr = "167.179.66.149:6379"

	addr, ok := os.LookupEnv("REDIS_ADDR")
	if !ok || addr == "" {
		return defaultAddr
	}

	return addr
}

func envRedisPort() string {
	const defaultPort = "6379"

	port, ok := os.LookupEnv("REDIS_PASS")
	if !ok || port == "" {
		return defaultPort
	}

	return port
}

func envRedisPassword() string {
	const defaultPass = "ooBae5ciZiCohf9L"

	pass, ok := os.LookupEnv("REDIS_PASS")
	if !ok || pass == "" {
		return defaultPass
	}

	return pass
}

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

func envPasswordKeyAES() string {
	const defaultKey = "my-secret-foobarbaz200030004000x"

	k, ok := os.LookupEnv("PASSWORD_KEY_AES")
	if !ok || len(k) < 32 {
		return defaultKey
	}

	return k
}

func mw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("==================")
		fmt.Println("from mw", r.URL)

		next.ServeHTTP(w, r)
	})
}

func main() {
	redisAddr := envRedisAddr()
	redisDb := envRedisDb()
	redisPass := envRedisPassword()
	passwordKey := envPasswordKeyAES()

	// เรียกใช้ redis
	repoClip := redisclipboard.New(redisAddr, redisDb, redisPass)
	repoUser := redisuser.New(redisAddr, redisDb, redisPass)
	servicePassword := service.NewServicePassword(passwordKey)

	// เรียก handler โดยที่ผ่านการเรียก redis ไปแล้ว
	hClip := handlerclipboard.NewClipboard(repoClip)
	hUser := handleruser.NewUser(repoUser, servicePassword)

	r := mux.NewRouter()

	r.Use(mw)

	r.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("from foo")
		fmt.Fprintf(w, "ok")
	})

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
	r.HandleFunc("/whoami", hUser.Whoami)

	err := http.ListenAndServe(":8000", r)
	if err != nil {
		log.Println("server error:", err)
	}
}
