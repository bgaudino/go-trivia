package handlers

import (
	"errors"
	"net/http"
	"time"
	"trivia/forms"
	"trivia/models"

	"github.com/google/uuid"
)

func loginRequired(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if models.GetSession(r) == nil {
			http.Redirect(w, r, "/admin/login", http.StatusTemporaryRedirect)
		} else {
			h(w, r)
		}
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	s := models.GetSession(r)
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
			expiry := time.Now().UTC().Add(48 * time.Hour)
			id := uuid.New()
			token := id.String()
			s := models.Session{
				Token:  token,
				Expiry: expiry,
				User:   &models.User{Username: user.Username, Id: user.Id},
			}
			s.Save()
			http.SetCookie(w, &http.Cookie{
				Name:    "session_token",
				Value:   token,
				Expires: expiry,
				Path:    "/",
			})
			http.Redirect(w, r, "/admin/", http.StatusTemporaryRedirect)
		}
	} else {
		Templates.ExecuteTemplate(w, "login.html", form)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	s := models.GetSession(r)
	if s != nil {
		s.Delete()
	}
	http.Redirect(w, r, "/admin/login/", http.StatusTemporaryRedirect)
}
