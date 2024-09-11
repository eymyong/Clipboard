package main

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

func main() {
	key := []byte("someKey")
	token, _, err := newJwt("clipboard-server", "userid-1", key)
	if err != nil {
		panic(err)
	}

	fmt.Println("token", token)

	payload, err := verifyJwt(token, key)
	if err != nil {
		panic(err)
	}

	fmt.Println("payload", payload)
}

func newJwt(iss, id string, key []byte) (token string, exp time.Time, err error) {
	// TODO: investigate if Local() is actually needed
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

func verifyJwt(tokenStr string, key []byte) (jwt.Claims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse JWT token %s", tokenStr)
	}
	return token.Claims, nil
}
