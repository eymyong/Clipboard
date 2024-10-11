package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	RedisAddr     string `json:"redisaddr"`
	RedisDb       int    `json:"redisdb"`
	RedisUsername string `json:"redisusername"`
	RedisPassword string `json:"redispassword"`
	// For encrypting password
	SecretAES string
	// For signing JWT
	SecretJWT string

	//DB
	DriverName string `json:"driverName"`
	DbHost     string `json:"dbhost"`
	DbPort     string `json:"dbport"`
	DbUser     string `json:"dbuser"`
	DbName     string `json:"dbname"`
}

func ReadJson(fileName string) Config {
	secretAes := "my-secret-foobarbaz200030004000x"
	secretJwt := "clipboard-jwt-secret"

	b, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	var env Config
	err = json.Unmarshal(b, &env)
	if err != nil {
		panic(err)
	}

	return Config{
		RedisAddr:     env.RedisAddr,
		RedisDb:       env.RedisDb,
		RedisUsername: env.RedisUsername,
		RedisPassword: env.RedisPassword,
		SecretAES:     secretAes,
		SecretJWT:     secretJwt,
		DbHost:        env.DbHost,
		DbPort:        env.DbPort,
		DbUser:        env.DbUser,
		DbName:        env.DbName,
		DriverName:    env.DriverName,
	}
}

// DataSourceName = "host=167.179.66.149 port=5469 user=postgres dbname=yongdb"
func DataSourceName(host, port, user, dbname string) string {
	c := ReadJson("config.json")
	dataSourceName := "host=" + c.DbHost + " post=" + c.DbPort + " user=" + c.DbUser + " dbname=" + c.DbName

	return dataSourceName
}

func EnvConfig(fileName string) Config {
	RedisAddr := "127.0.0.1:6379" //
	RedisDb := 0
	RedisUsername := ""
	RedisPassword := "" //

	secretAes := "my-secret-foobarbaz200030004000x"
	secretJwt := "clipboard-jwt-secret"

	redisDbEnvStr, _ := os.LookupEnv("REDIS_DB")
	// panic("here " + redisDbEnvStr)
	redisDbEnv, err := strconv.Atoi(redisDbEnvStr)
	if err != nil {
		fmt.Println("redisDbEnv err:%w", err.Error())
		panic("bad redis db config: " + redisDbEnvStr) //
	}

	RedisDb = redisDbEnv

	RedisAddrEnv, _ := os.LookupEnv("REDIS_ADDR")
	if RedisAddrEnv != "" {
		// RedisAddr = RedisAddrEnv //
	}

	redisUsernameEnv, _ := os.LookupEnv("REDIS_USERNAME")
	fmt.Println(RedisUsername)

	if redisUsernameEnv != "" {
		RedisUsername = redisUsernameEnv

	}

	redisPasswordEnv, _ := os.LookupEnv("REDIS_PASSWORD")
	if redisPasswordEnv != "" {
		//RedisAddr = RedisAddrEnv
	}

	secretAesEnv, _ := os.LookupEnv("SECRET_AES")
	if len(secretAesEnv) >= 32 {
		secretAes = secretAesEnv
	}

	secretJwtEnv, _ := os.LookupEnv("SECRET_JWT")
	if secretJwtEnv != "" {
		secretJwt = secretJwtEnv
	}

	return Config{
		RedisAddr:     RedisAddr,
		RedisDb:       RedisDb,
		RedisUsername: RedisUsername,
		RedisPassword: RedisPassword,
		SecretAES:     secretAes,
		SecretJWT:     secretJwt,
	}
}
