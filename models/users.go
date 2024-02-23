package models

import (
	"context"
	"fmt"
	"os"
	"trivia/db"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id       int
	Username string
	Password string
}

func GetUser(username string) (*User, error) {
	user := User{}
	row := db.Pool.QueryRow(context.Background(), "SELECT id, username, password FROM users WHERE username = $1", username)
	err := row.Scan(&user.Id, &user.Username, &user.Password)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		return nil, err
	}
	return &user, nil
}

func (u User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
