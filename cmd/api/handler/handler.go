package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/eymyong/drop/model"
	"github.com/eymyong/drop/repo"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type HandlerClip struct {
	repo repo.Repository
}

func New(repo repo.Repository) *HandlerClip {
	return &HandlerClip{repo: repo}
}

func sendJson(w http.ResponseWriter, status int, data interface{}) { //
	w.Header().Set("Content-Type", "application/json") // toi kumnood papet kormoon wa pen what? ex: json
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func readBody(r *http.Request) ([]byte, error) { //
	defer r.Body.Close()

	buf := bytes.NewBuffer(nil)
	_, err := io.Copy(buf, r.Body)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (h *HandlerClip) Create(w http.ResponseWriter, r *http.Request) {
	b, err := readBody(r)
	if err != nil {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "failed to read body",
			"reason": err.Error(),
		})
		return
	}

	clipboard := model.Clipboard{
		Id:   uuid.NewString(),
		Text: string(b),
	}

	err = h.repo.Create(nil, clipboard)
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "failed to create clipboard",
			"reason": err.Error(),
		})
		return
	}

	sendJson(w, http.StatusCreated, map[string]interface{}{
		"success": "ok",
		"created": clipboard,
	})
}

func (h *HandlerClip) GetAll(w http.ResponseWriter, r *http.Request) {
	clipboards, err := h.repo.GetAll(nil)
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "failed to get all todos",
			"reason": err.Error(),
		})

		return
	}

	sendJson(w, http.StatusOK, clipboards)
}

func (h *HandlerClip) GetById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)        //
	id := vars["clipboard-id"] //
	if id == "" {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error": "missing id",
		})

		return
	}

	clipboard, err := h.repo.GetById(nil, id)
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  fmt.Sprintf("failed to get todo %s", id),
			"reason": err.Error(),
		})

		return
	}

	sendJson(w, http.StatusOK, clipboard)

}
