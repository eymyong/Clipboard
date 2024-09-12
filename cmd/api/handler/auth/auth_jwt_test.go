package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eymyong/drop/cmd/api/handler/apiutils"
	"github.com/eymyong/drop/cmd/api/handler/auth"
)

func TestAuthJWT_InvalidToken(t *testing.T) {
	testSecret := "test-secret"
	authenticator := auth.New(testSecret)

	handler := func(w http.ResponseWriter, r *http.Request) {
		apiutils.SendJson(w, 200, nil)
	}

	next := http.HandlerFunc(handler)
	server := authenticator.AuthMiddlewareHeader(next)

	response := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodPost, "", nil)
	if err != nil {
		panic(err)
	}

	// Call API
	server.ServeHTTP(response, request)

	status := response.Result().StatusCode
	if status != http.StatusUnauthorized {
		t.Error("unexpected status", status)
	}
}

func TestAuthJWT_ValidToken(t *testing.T) {
	testSecret := "test-secret"
	testIss := "test-iss"
	testUserId := "test-user-id"

	authenticator := auth.New(testSecret)

	handler := func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get(auth.AuthHeaderUserId)
		if userID != testUserId {
			apiutils.SendJson(w, http.StatusInternalServerError, nil)
			return
		}

		apiutils.SendJson(w, http.StatusOK, nil)
	}

	next := http.HandlerFunc(handler)
	server := authenticator.AuthMiddlewareHeader(next)

	token, _, err := authenticator.NewTokenJWT(testIss, testUserId)
	if err != nil {
		panic(err)
	}

	response := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodPost, "", nil)
	if err != nil {
		panic(err)
	}

	request.Header.Set("Authorization", "Bearer "+token)

	// Call API
	server.ServeHTTP(response, request)

	status := response.Result().StatusCode
	if status != http.StatusOK {
		t.Error("unexpected status", status)
	}
}
