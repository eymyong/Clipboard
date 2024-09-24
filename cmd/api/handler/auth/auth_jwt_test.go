package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/eymyong/drop/cmd/api/handler/apiutils"
	"github.com/eymyong/drop/cmd/api/handler/auth"
)

func Test_NewtokenJWT(t *testing.T) {

	secretJwt := "clipboard-jwt-secret"
	expectedExp := time.Now().Add(24 * time.Hour).Local()
	expectedId := "123"
	expectedIss := "test"

	token, exp, err := auth.NewTokenJWT(expectedIss, expectedId, []byte(secretJwt))
	if err != nil {
		t.Errorf("unexpected err: %s", err.Error())
		return
	}

	if token == "" {
		t.Errorf("expect token but got: %s", token)
		return
	}

	if expectedExp != exp {
		t.Errorf("unexpected exp != expectedExp")
		return
	}

}

func Test_ExtractClaims(t *testing.T) {
	secretJwt := "clipboard-jwt-secret"
	expectedId := "123"
	expectedIss := "test"

	tokenStr, expectedExp, err := auth.NewTokenJWT(expectedIss, expectedId, []byte(secretJwt))
	if err != nil {
		t.Errorf("unexpected err: %s", err.Error())
		return
	}

	actualClaims, err := auth.ExtractClaims(tokenStr, []byte(secretJwt))
	if err != nil {
		t.Errorf("unexpected err: %s", err.Error())
	}

	c, ok := actualClaims.(jwt.MapClaims)
	if !ok {
		t.Errorf("unexpect ok: %v", ok)
	}

	_, ok = c["jti"]
	if !ok {
		t.Errorf("unexpected `jti` not found")
		return
	}

	id := c["jti"].(string)
	if id != expectedId {
		t.Errorf("expected id: %s but got id: %s", expectedId, id)
		return
	}

	_, ok = c["iss"]
	if !ok {
		t.Errorf("unexpected `iss` not found")
		return
	}

	issuer := c["iss"].(string)
	if issuer != expectedIss {
		t.Errorf("expected issuer: %s but got issuer: %s", expectedIss, issuer)
		return
	}

	_, ok = c["exp"]
	if !ok {
		t.Errorf("unexpected `exp` not found")
		return
	}

	exp := c["exp"].(float64)
	if exp != float64(expectedExp.Unix()) {
		t.Errorf("expected exp: %v but got exp: %v", expectedExp, exp)
		return
	}

}

func Test_AuthMiddlewareHeaderError(t *testing.T) {
	testSecret := "test-secert"
	authenticator := auth.New(testSecret)

	handler := func(w http.ResponseWriter, r *http.Request) {
		apiutils.SendJson(w, http.StatusOK, "ok")
	}

	next := http.HandlerFunc(handler)
	server := authenticator.AuthMiddlewareHeader(next)

	response := httptest.NewRecorder()
	request, err := http.NewRequest(http.MethodPost, "", nil)
	if err != nil {
		t.Errorf("unexpect request err:%s", err.Error())
	}

	//Call API
	server.ServeHTTP(response, request)

	status := response.Result().StatusCode
	if status != http.StatusUnauthorized {
		t.Error("unexpected status", status)
	}

}
func Test_AuthMiddlewareHeaderHappy(t *testing.T) {
	testSecret := "test-secret"
	testSecret2 := "test"
	testIss := "test-iss"
	testUserId := "test-user-id"
	authenticator := auth.New(testSecret)
	authenticator2 := auth.New(testSecret2)

	//
	handler := func(w http.ResponseWriter, r *http.Request) {
		userID := auth.GetUserIdFromHeader(r.Header)
		if userID != testUserId {
			apiutils.SendJson(w, http.StatusInternalServerError, "not found user-id")
			return
		}

		userIss := auth.GetUserIssFromHeader(r.Header)
		if userIss != testIss {
			apiutils.SendJson(w, http.StatusInternalServerError, "not found user-iss")
			return
		}

		apiutils.SendJson(w, http.StatusOK, "ok")
	}

	// create next send to `func AuthMiddlewareHeader``
	next := http.HandlerFunc(handler)
	server := authenticator.AuthMiddlewareHeader(next)

	// New response writer (add to `server.ServeHTTP`)
	response := httptest.NewRecorder()
	//newRequest (add to `server.ServeHTTP`)
	request, err := http.NewRequest(http.MethodPost, "", nil)
	if err != nil {
		t.Errorf("unexpected request err: %s", err.Error())
	}

	// create token แล้วsetใส่ r.Header เพื่อให้ `handler` get ออกมา
	// token, _, err := authenticator.NewTokenJWT(testIss, testUserId)
	// if err != nil {
	// 	t.Errorf("unexpected newToken err: %s", err.Error())
	// }

	token2, _, err := authenticator2.NewTokenJWT(testIss, testUserId)
	if err != nil {
		t.Errorf("unexpected newToken err: %s", err.Error())
	}

	request.Header.Set("Authorization", "Bearer "+token2)
	// request.Header.Set("Authorization", "Bearer "+token)

	// Call API
	server.ServeHTTP(response, request)

	status := response.Result().StatusCode
	if status != 401 {
		t.Error("unexpect status:", status)
	}

}

// mw -> handler

// TODO: test auth no token, invalid token
