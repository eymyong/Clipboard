package main

import (
	"net/http"

	"github.com/eymyong/drop/cmd/api/handler/handlerclipboard"
	"github.com/eymyong/drop/cmd/api/handler/handleruser"
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
	//h := handler.New(repoClip, repoUser)

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
	r.HandleFunc("/users/getall", hUser.GetAllUser).Methods(http.MethodGet)
	r.HandleFunc("/users/get/{user-id}", hUser.GetByIdUser).Methods(http.MethodGet)
	r.HandleFunc("/users/update/username/{user-id}", hUser.UpdateUserName).Methods(http.MethodPatch)
	r.HandleFunc("/users/update/password/{user-id}", hUser.UpdatePassword).Methods(http.MethodPatch)
	r.HandleFunc("/users/delete/{user-id}", hUser.DeleteUser).Methods(http.MethodDelete)
	r.HandleFunc("/users/deleteall", hUser.DeleteAllUser).Methods(http.MethodDelete)

	http.ListenAndServe(":8000", r)

}
