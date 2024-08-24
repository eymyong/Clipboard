package model

type Clipboard struct {
	Id   string
	Text string
}

type User struct {
	Name string
	Age  int
	//Sex       string
}

type Account struct {
	Id       string
	Username string
	Password string
}

type KeyAccount struct {
	Data map[string]string
}

// type _User struct {
// 	Id        string
// 	Username  string
// 	Password  string
// 	BirthDate time.Time
// }
