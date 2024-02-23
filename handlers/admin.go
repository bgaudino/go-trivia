package handlers

import "net/http"

func adminHandler(w http.ResponseWriter, r *http.Request) {
	Templates.ExecuteTemplate(w, "admin.html", nil)
}

var AdminHandler = loginRequired(adminHandler)
