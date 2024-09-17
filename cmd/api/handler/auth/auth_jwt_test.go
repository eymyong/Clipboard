package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgrijalva/jwt-go"

	"github.com/eymyong/drop/cmd/api/handler/apiutils"
	"github.com/eymyong/drop/cmd/api/handler/auth"
)

func TestJWT(t *testing.T) {
	t.Run("Creating and verifying JWT tokens", testJWT)
	t.Run("JWT middleware with invalid JWT token", testMw_InvalidTokenJWT)
	t.Run("JWT middleware with valid JWT token", testMw_ValidTokenJWT)
}

func testJWT(t *testing.T) {
	id, iss := "id", "iss"

	secretA := "secret-a"
	secretB := "secret-b"

	authA := auth.New(secretA)
	authB := auth.New(secretB)

	tokenA, _, err := authA.NewTokenJWT(iss, id)
	if err != nil {
		t.Errorf("unexpected error newTokenA: %s", err.Error())
	}

	tokenB, _, err := authB.NewTokenJWT(iss, id)
	if err != nil {
		t.Errorf("unexpected error newTokenB: %s", err.Error())
	}

	claimsA, err := authA.VerifyTokenJWT(tokenA)
	if err != nil {
		t.Errorf("unexpected error extractTokenA: %s", err.Error())
	}

	if claimsA == nil {
		t.Error("unexpected nil claims A")
	}

	idA, issA, _, err := auth.ExtractClaims(claimsA.(jwt.MapClaims))
	if err != nil {
		t.Error("unexpected error extract claims A")
	}

	if idA != id {
		t.Errorf("unexpected id for tokenA, expecting='%s', got='%s'", idA, id)
	}

	if issA != iss {
		t.Errorf("unexpected iss for tokenA, expecting='%s', got='%s'", issA, iss)
	}

	claimsB, err := authB.VerifyTokenJWT(tokenB)
	if err != nil {
		t.Errorf("unexpected error extractTokenB: %s", err.Error())
	}

	if claimsB == nil {
		t.Error("unexpected nil claims B")
	}

	idB, issB, _, err := auth.ExtractClaims(claimsB.(jwt.MapClaims))
	if err != nil {
		t.Errorf("unexpected error extract claims B")
	}

	if idB != id {
		t.Errorf("unexpected id for tokenA, expecting='%s', got='%s'", idA, id)
	}

	if issB != iss {
		t.Errorf("unexpected iss for tokenA, expecting='%s', got='%s'", issA, iss)
	}

	_, err = authB.VerifyTokenJWT(tokenA)
	if err == nil {
		t.Error("unexpected nil error authB.extractTokenA")
	}

	_, err = authA.VerifyTokenJWT(tokenB)
	if err == nil {
		t.Error("unexpected nil error authA.extractTokenB")
	}
}

func testMw_InvalidTokenJWT(t *testing.T) {
	testSecret := "test-secret"
	authenticator := auth.New(testSecret)

	handler := func(w http.ResponseWriter, r *http.Request) {
		apiutils.SendJson(w, http.StatusOK, "ok")
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

func testMw_ValidTokenJWT(t *testing.T) {
	testSecret := "test-secret"
	testIss := "test-iss"
	testUserId := "test-user-id"

	authenticator := auth.New(testSecret)

	handler := func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get(auth.AuthHeaderUserId)
		if userID != testUserId {
			apiutils.SendJson(w, http.StatusInternalServerError, "unauthorized: bad user-id")
			return
		}

		apiutils.SendJson(w, http.StatusOK, "ok")
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
