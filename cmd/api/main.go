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

	repoClip := initRepoClip()
	repoUser := initRepoUser()

	// hc := handler.New(repoClip, repoUser)
	// hc := handlerclipboard.NewClipboard(repoClip)
	h := handler.New(repoClip, repoUser)

	r := mux.NewRouter()
	r.HandleFunc("/clipboards/create", h.CreateClip).Methods(http.MethodPost)
	r.HandleFunc("/clipboards/getall", h.GetAllClips).Methods(http.MethodGet)
	r.HandleFunc("/clipboards/get/{clipboard-id}", h.GetClipById).Methods(http.MethodGet)
	r.HandleFunc("/clipboards/update/{clipboard-id}", h.UpdateClipById).Methods(http.MethodPatch)
	r.HandleFunc("/clipboards/delete/{clipboard-id}", h.DeleteClip).Methods(http.MethodDelete)

	r.HandleFunc("/users/register", h.Register).Methods(http.MethodPost)
	r.HandleFunc("/users/login", h.Login).Methods(http.MethodPost)
	r.HandleFunc("/users/getall", h.GetAllUser).Methods(http.MethodGet)
	r.HandleFunc("/users/get/{user-id}", h.GetByIdUser).Methods(http.MethodGet)
	r.HandleFunc("/users/update/username/{user-id}", h.UpdateUserName).Methods(http.MethodPatch)
	r.HandleFunc("/users/update/password/{user-id}", h.UpdatePassword).Methods(http.MethodPatch)
	r.HandleFunc("/users/delete/{user-id}", h.DeleteUser).Methods(http.MethodDelete)
	r.HandleFunc("/users/deleteall", h.DeleteAllUser).Methods(http.MethodDelete)

	http.ListenAndServe(":8000", r)

}
