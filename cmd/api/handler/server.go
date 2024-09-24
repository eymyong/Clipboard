package handler

import (
	"net/http"

	"github.com/eymyong/drop/cmd/api/handler/auth"
	"github.com/eymyong/drop/cmd/api/handler/handlerclipboard"
	"github.com/eymyong/drop/cmd/api/handler/handleruser"
	"github.com/gorilla/mux"
)

func NewServer(handlerUser *handleruser.HandlerUser, handlerClip *handlerclipboard.HandlerClipboard, auth *auth.AuthenticatorJWT) http.Handler {
	r := mux.NewRouter() // Main router

	routerClip := r.PathPrefix("/clipboards").Subrouter()
	routerUsers := r.PathPrefix("/users").Subrouter()

	r.HandleFunc("/register", handlerUser.Register).Methods(http.MethodPost)
	r.HandleFunc("/login", handlerUser.Login).Methods(http.MethodPost)

	// routerUsers.HandleFunc("/register", handlerUser.Register).Methods(http.MethodPost)
	// routerUsers.HandleFunc("/login", handlerUser.Login).Methods(http.MethodPost)

	routerClip.Use(auth.AuthMiddlewareHeader)
	routerUsers.Use(auth.AuthMiddlewareHeader)

	routerClip.HandleFunc("/create", handlerClip.CreateClip).Methods(http.MethodPost)
	routerClip.HandleFunc("/get-all", handlerClip.GetAllClips).Methods(http.MethodGet)
	routerClip.HandleFunc("/get/{clipboard-id}", handlerClip.GetClipById).Methods(http.MethodGet)
	routerClip.HandleFunc("/update/{clipboard-id}", handlerClip.UpdateClipById).Methods(http.MethodPatch)
	routerClip.HandleFunc("/delete/{clipboard-id}", handlerClip.DeleteClip).Methods(http.MethodDelete)
	routerClip.HandleFunc("/delete-all", handlerClip.DeleteAllClip).Methods(http.MethodDelete)

	routerUsers.HandleFunc("/get/{user-id}", handlerUser.GetUserById).Methods(http.MethodGet)
	routerUsers.HandleFunc("/update/username/{user-id}", handlerUser.UpdateUsername).Methods(http.MethodPatch)
	routerUsers.HandleFunc("/update/password/{user-id}", handlerUser.UpdatePassword).Methods(http.MethodPatch)
	routerUsers.HandleFunc("/delete/{user-id}", handlerUser.DeleteUser).Methods(http.MethodDelete)

	return r
}
