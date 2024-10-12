package handlerclipboard

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/eymyong/drop/cmd/api/handler/apiutils"
	"github.com/eymyong/drop/cmd/api/handler/auth"
	"github.com/eymyong/drop/model"
	"github.com/eymyong/drop/repo"
)

const (
	Database = "DB"
)

type HandlerClipboard struct {
	repoClipboard repo.RepositoryClipboard
}

func NewClipboard(repoClipboard repo.RepositoryClipboard) *HandlerClipboard {
	return &HandlerClipboard{
		repoClipboard: repoClipboard,
	}
}

func RegisterRoutesClipboardAPI(r *mux.Router, h *HandlerClipboard) {
	r.HandleFunc("/create", h.CreateClip).Methods(http.MethodPost)
	r.HandleFunc("/get-all", h.GetAllClips).Methods(http.MethodGet)
	r.HandleFunc("/get/{clipboard-id}", h.GetClipById).Methods(http.MethodGet)
	r.HandleFunc("/update/{clipboard-id}", h.UpdateClipById).Methods(http.MethodPatch)
	r.HandleFunc("/delete/{clipboard-id}", h.DeleteClip).Methods(http.MethodDelete)
	r.HandleFunc("/delete-all", h.DeleteAllClip).Methods(http.MethodDelete)
}

func (h *HandlerClipboard) CreateClip(w http.ResponseWriter, r *http.Request) {
	userId := auth.GetUserIdFromHeader(r.Header)
	if userId == "" {
		apiutils.SendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "missing user-id",
		})
		return
	}

	b, err := apiutils.ReadBody(r)
	if err != nil {
		apiutils.SendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "failed to read body",
			"reason": err.Error(),
		})
		return
	}

	if len(b) == 0 {
		apiutils.SendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error": "empty body",
		})
		return
	}

	clipboard := model.Clipboard{
		Id:     uuid.NewString(),
		UserId: userId,
		Text:   string(b),
	}
	ctx := context.Background()

	err = h.repoClipboard.Create(ctx, clipboard)
	if err != nil {
		apiutils.SendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "failed to create clipboard",
			"reason": err.Error(),
		})
		return
	}

	apiutils.SendJson(w, http.StatusCreated, map[string]interface{}{
		"success": "ok",
		"created": clipboard,
	})
}

func (h *HandlerClipboard) GetAllClips(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userId := auth.GetUserIdFromHeader(r.Header)
	if userId == "" {
		apiutils.SendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "failed to get userId",
		})
		return
	}

	clipboards, err := h.repoClipboard.GetAllUserClipboards(ctx, userId)
	if err != nil {
		apiutils.SendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "failed to get all todos",
			"reason": err.Error(),
		})
		return
	}

	apiutils.SendJson(w, http.StatusOK, clipboards)
}

func (h *HandlerClipboard) GetClipById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["clipboard-id"]
	if id == "" {
		apiutils.SendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error": "missing id",
		})
		return
	}

	ctx := context.Background()

	clipboard, err := h.repoClipboard.GetById(ctx, id)
	if err != nil {
		apiutils.SendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  fmt.Sprintf("failed to get todo %s", id),
			"reason": err.Error(),
		})
		return
	}

	apiutils.SendJson(w, http.StatusOK, clipboard)
}

func (h *HandlerClipboard) UpdateClipById(w http.ResponseWriter, r *http.Request) {
	b, err := apiutils.ReadBody(r)
	if err != nil {
		apiutils.SendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "failed to read body",
			"reason": err.Error(),
		})
		return
	}

	if len(b) == 0 {
		apiutils.SendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error": "empty body",
		})
		return
	}

	vars := mux.Vars(r)
	id := vars["clipboard-id"]
	if id == "" {
		apiutils.SendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error": "missing id",
		})
		return
	}

	ctx := context.Background()
	err = h.repoClipboard.Update(ctx, id, string(b))
	if err != nil {
		apiutils.SendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "failed to update",
			"reason": err.Error(),
		})
		return
	}

	apiutils.SendJson(w, http.StatusOK, map[string]interface{}{
		"sucess": fmt.Sprintf("update to id: %s", id),
		"reason": string(b),
	})

}

func (h *HandlerClipboard) DeleteClip(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["clipboard-id"]
	if id == "" {
		apiutils.SendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error": "missing id",
		})
		return
	}

	ctx := r.Context()
	err := h.repoClipboard.Delete(ctx, id)
	if err != nil {
		apiutils.SendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "failed to delete",
			"reason": err.Error(),
		})
		return
	}

	apiutils.SendJson(w, http.StatusOK, map[string]interface{}{
		"sucess": fmt.Sprintf("delete to id: %s", id),
	})
}

func (h *HandlerClipboard) DeleteAllClip(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := h.repoClipboard.DeleteAll(ctx)
	if err != nil {
		apiutils.SendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "failed to delete",
			"reason": err.Error(),
		})
		return
	}

	apiutils.SendJson(w, http.StatusOK, map[string]interface{}{
		"sucess": "ok",
	})
}
