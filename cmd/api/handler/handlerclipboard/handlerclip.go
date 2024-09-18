package handlerclipboard

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/eymyong/drop/cmd/api/handler/apiutils"
	"github.com/eymyong/drop/model"
	"github.com/eymyong/drop/repo"
)

type HandlerClipboard struct {
	repoClipboard repo.RepositoryClipboard
}

func NewClipboard(repoClipboard repo.RepositoryClipboard) *HandlerClipboard {
	return &HandlerClipboard{repoClipboard: repoClipboard}
}

func (h *HandlerClipboard) CreateClip(w http.ResponseWriter, r *http.Request) {
	userId := apiutils.GetUserId(r)
	if userId == "" {
		apiutils.SendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "failed to read userID",
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
	userId := apiutils.GetUserId(r)
	if userId == "" {
		apiutils.SendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "failed to read userID",
		})

		return
	}

	ctx := context.Background()
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
	userId := apiutils.GetUserId(r)
	if userId == "" {
		apiutils.SendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "failed to read userID",
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
	clipboard, err := h.repoClipboard.GetUserClipboard(ctx, id, userId)
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
	userId := apiutils.GetUserId(r)
	if userId == "" {
		apiutils.SendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "failed to read userID",
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

	newText := string(b)

	vars := mux.Vars(r)
	id := vars["clipboard-id"]
	if id == "" {
		apiutils.SendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error": "missing id",
		})
		return
	}

	ctx := context.Background()
	err = h.repoClipboard.UpdateUserClipboard(ctx, id, userId, newText)
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
	userId := apiutils.GetUserId(r)
	if userId == "" {
		apiutils.SendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "failed to read userID",
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
	err := h.repoClipboard.DeleteUserClipboard(ctx, id, userId)
	if err != nil {
		apiutils.SendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "failed to delete",
			"reason": err.Error(),
		})
		return
	}

	apiutils.SendJson(w, http.StatusOK, map[string]interface{}{
		"sucess": fmt.Sprintf("delete clipboard '%s'", id),
	})
}
