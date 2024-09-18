package handleruser

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/eymyong/drop/cmd/api/handler/apiutils"
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

// TODO ถ้ามี username แล้วจะ register ไม่ได้ //
func (h *HandlerUser) Register(w http.ResponseWriter, r *http.Request) {
	type requestRegister struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	b, err := apiutils.ReadBody(r)
	if err != nil {
		apiutils.SendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "failed to read body",
			"reason": err.Error(),
		})

		return
	}

	var req requestRegister
	err = json.Unmarshal(b, &req)
	if err != nil {
		apiutils.SendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "invalid body",
			"reason": err.Error(),
		})

		return
	}

	if req.Username == "" {
		apiutils.SendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "invalid body",
			"reason": "empty username",
		})

		return
	}

	if req.Password == "" {
		apiutils.SendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "invalid body",
			"reason": "empty password`",
		})

		return
	}

	password, err := h.servicePassword.EncryptBase64(req.Password)
	if err != nil {
		apiutils.SendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "failed to encrypt password",
			"reason": err.Error(),
		})

		return
	}

	user := model.User{
		Id:       uuid.NewString(),
		Username: req.Username,
		Password: password,
	}

	ctx := r.Context()
	_, err = h.repoUser.Create(ctx, user)
	if err != nil {
		apiutils.SendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "failed to register user",
			"reason": err.Error(),
		})

		return
	}

	apiutils.SendJson(w, http.StatusCreated, map[string]interface{}{
		"success": "successfully registered",
	})
}

func (h *HandlerUser) Login(w http.ResponseWriter, r *http.Request) {
	type requestLogin struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	b, err := apiutils.ReadBody(r)
	if err != nil {
		apiutils.SendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "failed to read body",
			"reason": err.Error(),
		})

		return
	}

	var req requestLogin
	err = json.Unmarshal(b, &req)
	if err != nil {
		apiutils.SendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "invalid body",
			"reason": err.Error(),
		})

		return
	}

	ctx := r.Context()
	passwordBase64, err := h.repoUser.GetPassword(ctx, req.Username)
	if err != nil {
		apiutils.SendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "login failed",
		})

		return
	}

	password, err := h.servicePassword.DecryptBase64(string(passwordBase64))
	if err != nil {
		apiutils.SendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "login failed",
		})

		return
	}

	if password != req.Password {
		apiutils.SendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "login failed",
		})

		return
	}

	id, err := h.repoUser.GetUserId(ctx, req.Username)
	if err != nil {
		apiutils.SendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "login failed",
			"reason": "failed to get userID",
		})

		return
	}

	token, exp, err := h.authenticator.NewTokenJWT("clipboard-login", id)
	if err != nil {
		apiutils.SendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "failed to create login token",
			"reason": err.Error(),
		})

		return
	}

	apiutils.SendJson(w, http.StatusOK, map[string]interface{}{
		"success":  "ok",
		"username": req.Username,
		"token":    token,
		"exp":      exp,
	})
}

func (h *HandlerUser) GetUserById(w http.ResponseWriter, r *http.Request) {
	apiutils.SendJson(w, 500, "not implemented")
}

func (h *HandlerUser) UpdateUsername(w http.ResponseWriter, r *http.Request) {
	apiutils.SendJson(w, 500, "not implemented")
}

func (h *HandlerUser) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	apiutils.SendJson(w, 500, "not implemented")
}

func (h *HandlerUser) DeleteUser(w http.ResponseWriter, r *http.Request) {
	apiutils.SendJson(w, 500, "not implemented")
}
