package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/eymyong/drop/model"
	"github.com/eymyong/drop/repo"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/soyart/gfc/pkg/gfc"
)

const (
	secretKey        = "my-secret-foobarbaz200030004000x"
	encodingPassword = gfc.EncodingBase64
)

type Handler struct {
	repoClipboard repo.RepositoryClipboard
	repoUser      repo.RepositoryUser
}

func New(repoClipboard repo.RepositoryClipboard, repoUser repo.RepositoryUser) *Handler {
	return &Handler{repoClipboard: repoClipboard, repoUser: repoUser}
}

func sendJson(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func readBody(r *http.Request) ([]byte, error) {
	defer r.Body.Close()

	buf := bytes.NewBuffer(nil)
	_, err := io.Copy(buf, r.Body)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (h *Handler) CreateClip(w http.ResponseWriter, r *http.Request) {
	b, err := readBody(r)
	if err != nil {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "failed to read body",
			"reason": err.Error(),
		})
		return
	}

	if len(b) == 0 {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error": "empty body",
		})
		return
	}

	clipboard := model.Clipboard{
		Id:   uuid.NewString(),
		Text: string(b),
	}
	ctx := context.Background()
	err = h.repoClipboard.Create(ctx, clipboard)
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

func (h *Handler) GetAllClips(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	clipboards, err := h.repoClipboard.GetAll(ctx)
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "failed to get all todos",
			"reason": err.Error(),
		})
		return
	}

	sendJson(w, http.StatusOK, clipboards)
}

func (h *Handler) GetClipById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["clipboard-id"]
	if id == "" {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error": "missing id",
		})
		return
	}
	ctx := context.Background()
	clipboard, err := h.repoClipboard.GetById(ctx, id)
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  fmt.Sprintf("failed to get todo %s", id),
			"reason": err.Error(),
		})
		return
	}

	sendJson(w, http.StatusOK, clipboard)
}

func (h *Handler) UpdateClipById(w http.ResponseWriter, r *http.Request) {
	b, err := readBody(r)
	if err != nil {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "failed to read body",
			"reason": err.Error(),
		})
		return
	}

	if len(b) == 0 {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error": "empty body",
		})
		return
	}

	vars := mux.Vars(r)
	id := vars["clipboard-id"]
	if id == "" {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error": "missing id",
		})
		return
	}

	ctx := context.Background()
	err = h.repoClipboard.Update(ctx, id, string(b))
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "failed to update",
			"reason": err.Error(),
		})
		return
	}

	sendJson(w, http.StatusOK, map[string]interface{}{
		"sucess": fmt.Sprintf("update to id: %s", id),
		"reason": string(b),
	})
}

func (h *Handler) DeleteClip(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["clipboard-id"]
	if id == "" {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error": "missing id",
		})
		return
	}
	ctx := context.Background()
	err := h.repoClipboard.Delete(ctx, id)
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "failed to delete",
			"reason": err.Error(),
		})
		return
	}

	sendJson(w, http.StatusOK, map[string]interface{}{
		"sucess": fmt.Sprintf("delete to id: %s", id),
	})
}

// TODO ถ้ามี username แล้วจะ register ไม่ได้ //
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	type requestRegister struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	b, err := readBody(r)
	if err != nil {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "failed to read body",
			"reason": err.Error(),
		})
	}

	var req requestRegister
	err = json.Unmarshal(b, &req)
	if err != nil {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "invalid body",
			"reason": err.Error(),
		})

		return
	}

	if req.Username == "" {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "invalid body",
			"reason": "empty username",
		})

		return
	}

	if req.Password == "" {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "invalid body",
			"reason": "empty password`",
		})

		return
	}

	password := bytes.NewBuffer([]byte(req.Password))
	ciphertext, err := gfc.EncryptGCM(password, []byte(secretKey))
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "failed to register",
		})

		return
	}

	ciphertextString, err := gfc.Encode(encodingPassword, ciphertext)
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "failed to register",
		})

		return
	}

	user := model.User{
		Id:       uuid.NewString(),
		Username: req.Username,
		Password: string(ciphertextString.Bytes()),
	}

	ctx := r.Context()
	_, err = h.repoUser.Create(ctx, user)
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "failed to register user",
			"reason": err.Error(),
		})

		return
	}

	sendJson(w, http.StatusOK, map[string]interface{}{
		"success": "successfully registered",
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	type requestLogin struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	b, err := readBody(r)
	if err != nil {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "failed to read body",
			"reason": err.Error(),
		})
	}

	var req requestLogin
	err = json.Unmarshal(b, &req)
	if err != nil {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "invalid body",
			"reason": err.Error(),
		})

		return
	}

	ctx := r.Context()
	pass, err := h.repoUser.GetPassword(ctx, req.Username)
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "login failed",
		})

		return
	}

	passwordString := bytes.NewBuffer(pass)
	ciphertext, err := gfc.Decode(encodingPassword, passwordString)
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "login failed",
		})

		return
	}

	password, err := gfc.DecryptGCM(ciphertext, []byte(secretKey))
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "login failed",
		})

		return
	}

	if string(password.Bytes()) != req.Password {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "login failed",
		})

		return
	}

	sendJson(w, http.StatusOK, map[string]interface{}{
		"success": "hello " + req.Username,
	})
}

func (h *Handler) GetAllUser(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	users, err := h.repoUser.GetAll(ctx)
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "failed to get all users",
			"reason": err.Error(),
		})
		return
	}

	sendJson(w, http.StatusOK, users)
}

func (h *Handler) GetByIdUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["user-id"]
	if id == "" {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error": "missing id",
		})
		return
	}
	ctx := context.Background()
	user, err := h.repoUser.GetById(ctx, id)
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  fmt.Sprintf("failed to get user: %s", id),
			"reason": err.Error(),
		})
		return
	}

	sendJson(w, http.StatusOK, map[string]interface{}{
		"get by id": id,
		"reson":     user,
	})
}

func (h *Handler) UpdateUserName(w http.ResponseWriter, r *http.Request) {
	b, err := readBody(r)
	if err != nil {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "failed to read body",
			"reason": err.Error(),
		})
		return
	}

	if len(b) == 0 {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error": "empty body",
		})
		return
	}

	vars := mux.Vars(r)
	id := vars["user-id"]
	if id == "" {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error": "missing id",
		})
		return
	}

	ctx := r.Context()
	err = h.repoUser.UpdateUserName(ctx, id, string(b))
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "failed to update user",
			"reason": err.Error(),
		})
		return
	}

	sendJson(w, http.StatusOK, map[string]interface{}{
		"sucess": fmt.Sprintf("update to id: %s", id),
		"reason": string(b),
	})
}

func (h *Handler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	type requestUpdatePassword struct {
		Username    string `json:username`
		Password    string `json:password`
		NewPassword string `json:newpassword`
	}

	dataByte, err := readBody(r)
	if err != nil {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "failed to read body",
			"reason": err.Error(),
		})
		return
	}

	vars := mux.Vars(r)
	id := vars["user-id"]
	if id == "" {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error": "missing id",
		})
		return
	}

	var req requestUpdatePassword
	err = json.Unmarshal(dataByte, &req)
	if err != nil {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "invalid body",
			"reason": err.Error(),
		})
		return
	}

	if req.NewPassword == req.Password {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error": "expected new password",
		})
		return
	}

	ctx := r.Context()
	passByte, err := h.repoUser.GetPassword(ctx, req.Username)
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "failed to update password",
			"reason": err.Error(),
		})
		return
	}

	passwordString := bytes.NewBuffer(passByte)
	ciphertext, err := gfc.Decode(encodingPassword, passwordString)
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "failed to update password",
		})
		return
	}

	password, err := gfc.DecryptGCM(ciphertext, []byte(secretKey))
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "failed to update password",
		})
		return
	}

	if string(password.Bytes()) != req.Password {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "failed to update password",
			"reson": "password is incorrect",
		})
		return
	}

	newPassword := bytes.NewBuffer([]byte(req.NewPassword))
	newCiphertext, err := gfc.EncryptGCM(newPassword, []byte(secretKey))
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "failed to update password",
		})
		return
	}

	newCiphertextString, err := gfc.Encode(encodingPassword, newCiphertext)
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "failed to update password",
		})
		return
	}

	err = h.repoUser.UpdatePassword(ctx, id, string(newCiphertextString.Bytes()))
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "failed to update password",
			"reason": err.Error(),
		})
		return
	}

	sendJson(w, http.StatusOK, map[string]interface{}{
		"success": "update password to id: " + id,
	})
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["user-id"]
	if id == "" {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error": "missing id",
		})
		return
	}
	ctx := context.Background()
	err := h.repoUser.Delete(ctx, id)
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "failed to delete",
			"reason": err.Error(),
		})
		return
	}

	sendJson(w, http.StatusOK, map[string]interface{}{
		"sucess": fmt.Sprintf("delete to id: %s", id),
	})
}

func (h *Handler) DeleteAllUser(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	err := h.repoUser.DeleteAll(ctx)
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "failed to deleteall",
			"reson": err.Error(),
		})
	}

	sendJson(w, http.StatusOK, map[string]interface{}{
		"sucess": "delete all in redis [selecct 3]",
	})
}
