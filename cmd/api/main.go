package main

import (
	"net/http"

	"github.com/eymyong/drop/cmd/api/handler"
	"github.com/eymyong/drop/repo"
	"github.com/eymyong/drop/repo/redisclipboardwindows"
	"github.com/eymyong/drop/repo/redisuser"
	"github.com/gorilla/mux"
)

const RedisClipboardLinux = "linux"
const RedisClipboardWindows = "windows"
const RedisUser = "user"

func initRepoClip() repo.RepositoryClipboard {
	repo := redisclipboardwindows.New("127.0.0.1:6379", 2)
	return repo
}

func initRepoUser() repo.RepositoryUser {
	repo := redisuser.New("127.0.0.1:6379", 3)
	return repo
}

func main() {

	repoUser := initRepoUser()
	repoClip := initRepoClip()

	h := handler.New(repoClip, repoUser)

	r := mux.NewRouter()
	r.HandleFunc("/create", h.CreateClip).Methods(http.MethodPost)
	r.HandleFunc("/get-a", h.GetAllClips).Methods(http.MethodGet)
	r.HandleFunc("/getById/{clipboard-id}", h.GetClipById).Methods(http.MethodGet)
	r.HandleFunc("/update/{clipboard-id}", h.UpdateClipById).Methods(http.MethodPatch)
	r.HandleFunc("/delete/{clipboard-id}", h.DeleteClip).Methods(http.MethodDelete)
	r.HandleFunc("/register/{user-name}/{age}", h.Register).Methods(http.MethodPost)
	r.HandleFunc("/login/{user-name}", h.Login).Methods(http.MethodGet)

	http.ListenAndServe(":8000", r)

}
