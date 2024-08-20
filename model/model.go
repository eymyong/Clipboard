package model

type Clipboard struct {
	Id   string
	Text string
}

type User struct {
	Firstname string
	Lastname  string
	Age       int
	Sex       string
}

type Login struct {
	Username string
	Password string
}
