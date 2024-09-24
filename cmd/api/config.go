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
	//redisAddr := "127.0.0.1:6379" //
	redisDb := 0
	redisUsername := ""
	//redisPassword := "" //
	secretAes := "my-secret-foobarbaz200030004000x"
	secretJwt := "clipboard-jwt-secret"

	redisDbEnvStr, _ := os.LookupEnv("REDIS_DB")
	redisDbEnv, err := strconv.Atoi(redisDbEnvStr)
	if err != nil {
		panic("bad redis db config: " + redisDbEnvStr) //
	}

	redisDb = redisDbEnv

	redisAddrEnv, _ := os.LookupEnv("REDIS_ADDR")
	if redisAddrEnv != "" {
		//redisAddr = redisAddrEnv //
	}

	redisUsernameEnv, _ := os.LookupEnv("REDIS_USERNAME")
	if redisUsernameEnv != "" {
		redisUsername = redisUsernameEnv
	}

	redisPasswordEnv, _ := os.LookupEnv("REDIS_PASSWORD")
	if redisPasswordEnv != "" {
		//redisAddr = redisAddrEnv
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
		redisAddr:     "167.179.66.149:6379",
		redisDb:       redisDb,
		redisUsername: redisUsername,
		redisPassword: "Eepi2geeque2ahCo",
		secretAES:     secretAes,
		secretJWT:     secretJwt,
	}
}
