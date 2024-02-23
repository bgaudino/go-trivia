package models

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
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
		fmt.Fprintln(os.Stderr, err.Error())
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

type Session struct {
	Id     int
	Token  string
	User   *User
	Expiry time.Time
}

func (s *Session) IsExpired() bool {
	return s.Expiry.Before(time.Now().UTC())
}

func GetSession(r *http.Request) *Session {
	c, err := r.Cookie("session_token")
	if err != nil {
		return nil
	}
	session := Session{
		User: &User{},
	}
	row := db.Pool.QueryRow(context.Background(),
		`
			SELECT sessions.id, sessions.expiry, users.id, users.username
			FROM sessions
			JOIN users ON sessions.user_id = users.id
			WHERE sessions.token = $1
		`,
		c.Value,
	)
	err = row.Scan(&session.Id, &session.Expiry, &session.User.Id, &session.User.Username)
	if err != nil {
		return nil
	}
	if session.IsExpired() {
		session.Delete()
		return nil
	}
	return &session
}

func (s *Session) Save() error {
	row := db.Pool.QueryRow(
		context.Background(),
		"INSERT INTO sessions (token, expiry, user_id) VALUES ($1, $2, $3) RETURNING id",
		s.Token, s.Expiry, s.User.Id,
	)
	return row.Scan(&s.Id)
}

func (s *Session) Delete() {
	db.Pool.Exec(context.Background(), "DELETE FROM sessions WHERE id = $1", s.Id)
}
