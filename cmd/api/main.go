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
	r.HandleFunc("/clipboard", h.CreateClip).Methods(http.MethodPost)
	r.HandleFunc("/clipboards/all", h.GetAllClips).Methods(http.MethodGet)
	r.HandleFunc("/clipboards/{clipboard-id}", h.GetClipById).Methods(http.MethodGet)
	r.HandleFunc("/clipboards/update/{clipboard-id}", h.UpdateClipById).Methods(http.MethodPatch)
	r.HandleFunc("/clipboards/delete/{clipboard-id}", h.DeleteClip).Methods(http.MethodDelete)

	r.HandleFunc("/users/register", h.Register).Methods(http.MethodPost)
	r.HandleFunc("/users/login", h.Login).Methods(http.MethodPost)
	r.HandleFunc("/users/getall", h.GetAllUser).Methods(http.MethodGet)

	http.ListenAndServe(":8000", r)

}
