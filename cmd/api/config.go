package main

import (
	"os"
	"strconv"
)

type config struct {
	redisAddr     string
	redisDb       int
	redisUsername string
	redisPassword string

	secretAES string // For encrypting password
	secretJWT string // For signing JWT
}

func envConfig() config {
	redisAddr := "127.0.0.1:6379"
	redisDb := 0
	redisUsername := ""
	redisPassword := ""
	secretAes := "my-secret-foobarbaz200030004000x"
	secretJwt := "clipboard-jwt-secret"

	redisDbEnvStr, _ := os.LookupEnv("REDIS_DB")
	redisDbEnv, err := strconv.Atoi(redisDbEnvStr)
	if err != nil {
		redisDbEnv = 0
	}

	redisDb = redisDbEnv

	redisAddrEnv, _ := os.LookupEnv("REDIS_ADDR")
	if redisAddrEnv != "" {
		redisAddr = redisAddrEnv
	}

	redisUsernameEnv, _ := os.LookupEnv("REDIS_USERNAME")
	if redisUsernameEnv != "" {
		redisUsername = redisUsernameEnv
	}

	redisPasswordEnv, _ := os.LookupEnv("REDIS_PASSWORD")
	if redisPasswordEnv != "" {
		redisPassword = redisPasswordEnv
	}

	secretAesEnv, _ := os.LookupEnv("SECRET_AES")
	if len(secretAesEnv) >= 32 {
		secretAes = secretAesEnv
	}

	secretJwtEnv, _ := os.LookupEnv("SECRET_JWT")
	if secretJwtEnv != "" {
		secretJwt = secretJwtEnv
	}

	return config{
		redisAddr:     redisAddr,
		redisDb:       redisDb,
		redisUsername: redisUsername,
		redisPassword: redisPassword,
		secretAES:     secretAes,
		secretJWT:     secretJwt,
	}
}
