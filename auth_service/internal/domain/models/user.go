package models

type User struct {
	ID       int64
	Username string
	Password []byte //in db hashed pass
}
