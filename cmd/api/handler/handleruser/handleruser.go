package handleruser

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"

	"github.com/eymyong/drop/cmd/api/handler/auth"
	"github.com/eymyong/drop/cmd/api/service"
	"github.com/eymyong/drop/model"
	"github.com/eymyong/drop/repo"
)

type HandlerUser struct {
	repoUser        repo.RepositoryUser
	servicePassword service.Password
	authenticator   auth.Authenticator
}

func NewUser(
	repoUser repo.RepositoryUser,
	servicePassword service.Password,
	authenticator auth.Authenticator,
) *HandlerUser {
	return &HandlerUser{
		repoUser:        repoUser,
		servicePassword: servicePassword,
		authenticator:   authenticator,
	}
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

// TODO ถ้ามี username แล้วจะ register ไม่ได้ //
func (h *HandlerUser) Register(w http.ResponseWriter, r *http.Request) {
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

	password, err := h.servicePassword.EncryptBase64(req.Password)
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "failed to encrypt password",
			"reason": err.Error(),
		})
	}

	user := model.User{
		Id:       uuid.NewString(),
		Username: req.Username,
		Password: password,
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

	sendJson(w, http.StatusCreated, map[string]interface{}{
		"success": "successfully registered",
	})
}

func (h *HandlerUser) Login(w http.ResponseWriter, r *http.Request) {
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

		return
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
	passwordBase64, err := h.repoUser.GetPassword(ctx, req.Username)
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "login failed",
		})

		return
	}

	password, err := h.servicePassword.DecryptBase64(string(passwordBase64))
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "login failed",
		})

		return
	}

	if password != req.Password {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "login failed",
		})

		return
	}

	id, err := h.repoUser.GetUserId(ctx, req.Username)
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "login failed",
			"reason": "failed to get userID",
		})

		return
	}

	token, exp, err := h.authenticator.NewTokenJWT("clipboard-login", id)
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "failed to create login token",
			"reason": err.Error(),
		})
	}

	sendJson(w, http.StatusOK, map[string]interface{}{
		"success":  "ok",
		"username": req.Username,
		"token":    token,
		"exp":      exp,
	})
}

func (h *HandlerUser) GetUserById(w http.ResponseWriter, r *http.Request) {
	sendJson(w, 500, "not implemented")
}

func (h *HandlerUser) UpdateUsername(w http.ResponseWriter, r *http.Request) {
	sendJson(w, 500, "not implemented")
}

func (h *HandlerUser) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	sendJson(w, 500, "not implemented")
}

func (h *HandlerUser) DeleteUser(w http.ResponseWriter, r *http.Request) {
	sendJson(w, 500, "not implemented")
}
