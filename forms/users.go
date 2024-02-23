package forms

import (
	"errors"
	"net/http"
	"strings"
	"trivia/models"
)

type UserForm struct {
	Request *http.Request
	Errors  map[string][]error
	Model   *models.User
}

func NewUserForm(r *http.Request, m *models.User) UserForm {
	f := UserForm{Request: r, Model: m, Errors: make(map[string][]error)}
	if f.Model == nil {
		f.Model = &models.User{}
	}
	return f
}

func (f *UserForm) IsValid() bool {
	f.Request.ParseForm()
	username := strings.TrimSpace(f.Request.Form.Get("username"))
	password := strings.TrimSpace(f.Request.Form.Get("password"))
	if username == "" {
		f.Errors["username"] = []error{errors.New("this field is required")}
	} else {
		f.Model.Username = username
	}
	if password == "" {
		f.Errors["password"] = []error{errors.New("this field is required")}
	} else {
		f.Model.Password = password
	}
	return len(f.Errors) == 0
}
