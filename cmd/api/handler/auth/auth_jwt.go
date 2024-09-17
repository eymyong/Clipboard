package auth

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/pkg/errors"
)

const (
	AuthHeaderUserId = "drop-user-id"
	AuthHeaderIssuer = "drop-user-iss"
	AuthHeaderExpiry = "drop-user-exp"
)

type Authenticator interface {
	NewTokenJWT(iss, id string) (token string, exp time.Time, err error)
}

type AuthenticatorJWT struct {
	secretKey string
}

func New(secretKey string) *AuthenticatorJWT {
	return &AuthenticatorJWT{
		secretKey: secretKey,
	}
}

// For jwt embedded in request header
func (a *AuthenticatorJWT) AuthMiddlewareHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if len(tokenStr) == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Missing Authorization Header"))
			return
		}

		tokenStr = strings.Replace(tokenStr, "Bearer ", "", 1)
		claims, err := ExtractClaims(tokenStr, []byte(a.secretKey))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("error verifying JWT token: " + err.Error()))
			return
		}

		c, ok := claims.(jwt.MapClaims)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "unexpected jwt type: '%s', expecting '%s'", reflect.TypeOf(claims).String(), reflect.TypeOf(c).String())
			return
		}

		idClaims, ok := c["jti"]
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "null value at key jti")
			return
		}

		id, ok := idClaims.(string)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "unexpected jwt jti: '%s', expecting '%s'", reflect.TypeOf(idClaims).String(), reflect.TypeOf(id).String())
			return
		}

		issClaims, ok := c["iss"]
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "null value at key jti")
			return
		}

		iss, ok := issClaims.(string)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "unexpected jwt iss: '%s', expecting '%s'", reflect.TypeOf(issClaims).String(), reflect.TypeOf(iss).String())
			return
		}

		expClaims, ok := c["exp"]
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "null value at key exp")
			return
		}

		exp, ok := expClaims.(float64)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "unexpected jwt exp: '%s', expecting '%s'", reflect.TypeOf(expClaims).String(), reflect.TypeOf(exp).String())
			return
		}

		r.Header.Set(AuthHeaderUserId, id)
		r.Header.Set(AuthHeaderIssuer, iss)
		r.Header.Set(AuthHeaderExpiry, fmt.Sprintf("%f", exp))

		next.ServeHTTP(w, r)
	})
}

func (a *AuthenticatorJWT) NewTokenJWT(iss, id string) (token string, exp time.Time, err error) {
	return NewTokenJWT(iss, id, []byte(a.secretKey))
}

func NewTokenJWT(iss, id string, key []byte) (token string, exp time.Time, err error) {
	exp = time.Now().Add(24 * time.Hour).Local()
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Id:        id,
		Issuer:    iss,
		ExpiresAt: exp.Unix(),
	})

	// Generate JWT token from claims
	token, err = claims.SignedString(key)
	if err != nil {
		return token, exp, errors.Wrapf(err, "failed to validate with key %s", key)
	}

	return token, exp, nil
}

func ExtractClaims(tokenStr string, key []byte) (jwt.Claims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse JWT token %s", tokenStr)
	}

	return token.Claims, nil
}
