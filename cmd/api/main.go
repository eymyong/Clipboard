package main

import (
	"net/http"

	"github.com/eymyong/drop/cmd/api/handler"
	"github.com/eymyong/drop/repo/redisclipboard"
	"github.com/gorilla/mux"
)

func main() {
	repo := redisclipboard.New("127.0.0.1:6379")
	h := handler.New(repo)

	r := mux.NewRouter()
	r.HandleFunc("/create", h.Create).Methods(http.MethodGet)
	r.HandleFunc("/getAll", h.GetAll).Methods(http.MethodGet)
	r.HandleFunc("/getById/{clipboard-id}", h.GetById).Methods(http.MethodGet)

	http.ListenAndServe(":8000", r)

}
