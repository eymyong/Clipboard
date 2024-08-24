package model

type Clipboard struct {
	Id   string
	Text string
}

type User struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}
