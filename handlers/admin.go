package handlers

import (
	"net/http"
	"trivia/models"
)

func adminHandler(w http.ResponseWriter, r *http.Request) {
	session := models.GetSession(r)
	Templates.ExecuteTemplate(w, "admin.html", session.User)
}

var AdminHandler = loginRequired(adminHandler)
