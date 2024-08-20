package main

import (
	"net/http"
	"os"

	"github.com/eymyong/drop/cmd/api/handler"
	"github.com/eymyong/drop/repo"
	"github.com/eymyong/drop/repo/redisclipboardlinux"
	"github.com/eymyong/drop/repo/redisclipboardwindows"
	"github.com/gorilla/mux"
)

const RedisLinux = "linux"
const RedisWindows = "windows"

func initRepo() repo.Repository {
	envRepo := os.Getenv("REPO")

	var repo repo.Repository
	switch envRepo {
	case RedisLinux:
		repo = redisclipboardlinux.New("127.0.0.1:6379", 1)
	case RedisWindows:
		repo = redisclipboardwindows.New("127.0.0.1:6379", 2)
	}

	return repo
}

func main() {
	repo := initRepo()
	h := handler.New(repo)

	r := mux.NewRouter()
	r.HandleFunc("/create", h.Create).Methods(http.MethodPost)
	r.HandleFunc("/get-a", h.GetAll).Methods(http.MethodGet)
	r.HandleFunc("/getById/{clipboard-id}", h.GetById).Methods(http.MethodGet)
	r.HandleFunc("/update/{clipboard-id}", h.Update).Methods(http.MethodPatch)
	r.HandleFunc("/delete/{clipboard-id}", h.Delete).Methods(http.MethodDelete)

	http.ListenAndServe(":8000", r)

}
