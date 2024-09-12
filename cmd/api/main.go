package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/eymyong/drop/cmd/api/handler/auth"
	"github.com/eymyong/drop/cmd/api/handler/handlerclipboard"
	"github.com/eymyong/drop/cmd/api/handler/handleruser"
	"github.com/eymyong/drop/cmd/api/service"
	"github.com/eymyong/drop/repo"
	"github.com/eymyong/drop/repo/redisclipboard"
	"github.com/eymyong/drop/repo/redisuser"
)

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
	conf := envConfig()

	rd := repo.NewRedis(conf.redisAddr, conf.redisUsername, conf.redisPassword, conf.redisDb)
	repoClip := redisclipboard.New(rd)
	repoUser := redisuser.New(rd)

	servicePassword := service.NewServicePassword(conf.secretAES)
	authenticator := auth.New(conf.secretJWT)

	handlerClip := handlerclipboard.NewClipboard(repoClip)
	handlerUser := handleruser.NewUser(repoUser, servicePassword, authenticator)

	r := mux.NewRouter()               // Main router
	j := r.Path("/whoami").Subrouter() // Router for testing JWT

	j.Use(authenticator.AuthMiddlewareBody)
	j.HandleFunc("", whoAmI)

	r.HandleFunc("/clipboards/create", handlerClip.CreateClip).Methods(http.MethodPost)
	r.HandleFunc("/clipboards/get-all", handlerClip.GetAllClips).Methods(http.MethodGet)
	r.HandleFunc("/clipboards/get/{clipboard-id}", handlerClip.GetClipById).Methods(http.MethodGet)
	r.HandleFunc("/clipboards/update/{clipboard-id}", handlerClip.UpdateClipById).Methods(http.MethodPatch)
	r.HandleFunc("/clipboards/delete/{clipboard-id}", handlerClip.DeleteClip).Methods(http.MethodDelete)

	r.HandleFunc("/users/register", handlerUser.Register).Methods(http.MethodPost)
	r.HandleFunc("/users/login", handlerUser.Login).Methods(http.MethodPost)
	r.HandleFunc("/users/get/{user-id}", handlerUser.GetUserById).Methods(http.MethodGet)
	r.HandleFunc("/users/update/username/{user-id}", handlerUser.UpdateUsername).Methods(http.MethodPatch)
	r.HandleFunc("/users/update/password/{user-id}", handlerUser.UpdatePassword).Methods(http.MethodPatch)
	r.HandleFunc("/users/delete/{user-id}", handlerUser.DeleteUser).Methods(http.MethodDelete)

	err := http.ListenAndServe(":8000", r)
	if err != nil {
		log.Println("server error:", err)
	}
}
