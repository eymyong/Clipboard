package handleruser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/eymyong/drop/cmd/api/service"
	"github.com/eymyong/drop/model"
	"github.com/eymyong/drop/repo"
	"github.com/eymyong/drop/repo/redisuser"
)

type HandlerUser struct {
	repoUser        repo.RepositoryUser
	servicePassword service.Password
}

func NewUser(repoUser repo.RepositoryUser, servicePassword service.Password) *HandlerUser {
	return &HandlerUser{
		repoUser:        repoUser,
		servicePassword: servicePassword,
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
			"error":  "login failed",
			"reason": err.Error(),
		})

		return
	}

	password, err := h.servicePassword.DecryptBase64(string(passwordBase64))
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "login failed",
			"reason": err.Error(),
		})

		return
	}

	if password != req.Password {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "login failed",
			"reason": err.Error(),
		})

		return
	}
	// implemented create jwt

	id, err := h.repoUser.GetByUsername(ctx, req.Username)
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			// ควรแจ้งยังไงง //
			"error": "login failed",
			"reson": err.Error(),
		})

		return
	}

	token, err := redisuser.NewJwtTest(id, req.Username, []byte(redisuser.JwtKey))
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			// ควรแจ้งยังไงง //
			"error":  "create token failed",
			"reason": err.Error(),
		})

		return
	}

	sendJson(w, http.StatusOK, map[string]interface{}{
		"success":  "ok",
		"username": req.Username,
		"token":    token,
	})
}

func (h *HandlerUser) GetUserById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["user-id"]
	if id == "" {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error": "empty id",
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
		"success": "ok",
		"user":    user,
	})
}

func (h *HandlerUser) UpdateUsername(w http.ResponseWriter, r *http.Request) {
	b, err := readBody(r)
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
			"error": "missing userId",
		})

		return
	}

	if len(b) == 0 {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error": "empty body",
		})
		return
	}

	newUsername := string(b)
	ctx := r.Context()

	err = h.repoUser.UpdateUsername(ctx, id, newUsername)
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  fmt.Sprintf("failed to update userId '%s'", id),
			"reason": err.Error(),
		})

		return
	}

	sendJson(w, http.StatusOK, map[string]interface{}{
		"sucess": fmt.Sprintf("user id '%s' username updated to '%s'", id, newUsername),
	})
}

func (h *HandlerUser) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	sendJson(w, 500, "not implemented")
}

func (h *HandlerUser) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["user-id"]
	if id == "" {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error": "missing userId",
		})

		return
	}

	ctx := context.Background()
	err := h.repoUser.Delete(ctx, id)
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  fmt.Sprintf("failed to delete userId '%s'", id),
			"reason": err.Error(),
		})

		return
	}

	sendJson(w, http.StatusOK, "deleted userId: "+id)
}

func (h *HandlerUser) Whoami(w http.ResponseWriter, r *http.Request) {
	dataByte, err := readBody(r)
	if err != nil {
		sendJson(w, http.StatusBadRequest, map[string]interface{}{
			"error": "readbody err",
			"reson": err.Error(),
		})
	}

	token := string(dataByte)

	claims, err := redisuser.VerifyJwt(token, []byte(redisuser.JwtKey))
	if err != nil {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		})
	}

	if mapClaims, ok := claims.(jwt.MapClaims); ok {
		id := mapClaims["jti"]
		sendJson(w, http.StatusOK, id)

	} else {
		sendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "claims is not JWT mapClaims",
			"claims": claims,
		})
	}

}
