package handlers

import (
	"errors"
	"net/http"
	"time"
	"trivia/forms"
	"trivia/models"

	"github.com/google/uuid"
)

type session struct {
	username string
	expiry   time.Time
}

func (s *session) IsExpired() bool {
	return s.expiry.Before(time.Now())
}

var sessions = map[string]session{}

func getSession(r *http.Request) *session {
	c, err := r.Cookie("session_token")
	if err != nil {
		return nil
	}
	s, exists := sessions[c.Value]
	if !exists {
		return nil
	}
	if s.IsExpired() {
		delete(sessions, c.Value)
		return nil
	}
	return &s
}

func loginRequired(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if getSession(r) == nil {
			http.Redirect(w, r, "/admin/login", http.StatusTemporaryRedirect)
		} else {
			h(w, r)
		}
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	s := getSession(r)
	if s != nil {
		http.Redirect(w, r, "/admin/", http.StatusTemporaryRedirect)
		return
	}
	form := forms.NewUserForm(r, nil)
	if r.Method == "POST" {
		r.ParseForm()
		if !form.IsValid() {
			Templates.ExecuteTemplate(w, "login.html", form)
			return
		}
		user, err := models.GetUser(form.Model.Username)
		if err != nil || !user.CheckPassword(form.Model.Password) {
			form.Errors["_nonFieldErrors"] = []error{errors.New("invalid username or password")}
			Templates.ExecuteTemplate(w, "login.html", form)
			return
		} else {
			expiry := time.Now().Add(120 * time.Second)
			id := uuid.New()
			token := id.String()
			sessions[token] = session{
				username: user.Username,
				expiry:   time.Now().Add(120 * time.Second),
			}
			http.SetCookie(w, &http.Cookie{
				Name:    "session_token",
				Value:   token,
				Expires: expiry,
				Path:    "/",
			})
			r.Method = "GET"
			http.Redirect(w, r, "/admin/", http.StatusTemporaryRedirect)
		}
	} else {
		Templates.ExecuteTemplate(w, "login.html", form)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_token")
	if err == nil {
		delete(sessions, c.Value)
	}
	http.Redirect(w, r, "/admin/login/", http.StatusTemporaryRedirect)
}
