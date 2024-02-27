package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"trivia/db"
	"trivia/handlers"

	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}

	handlers.Templates = handlers.GetTemplates()
	r := http.NewServeMux()

	db.Pool, err = db.GetPool()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		log.Fatal("Unable to connect to database")
	}
	defer db.Pool.Close()

	r.HandleFunc("/", handlers.OptionsHandler)
	r.HandleFunc("/play/", handlers.PlayHandler)
	r.HandleFunc("/api/answer/", handlers.AnswerHandler)
	r.HandleFunc("/admin/", handlers.AdminHandler)
	r.HandleFunc("/admin/questions/add/", handlers.QuestionFormHandler)
	r.HandleFunc("/admin/login/", handlers.Login)
	r.HandleFunc("/admin/logout/", handlers.Logout)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	key := []byte(os.Getenv("SECRET_KEY"))
	log.Fatal(http.ListenAndServe(":"+port, csrf.Protect(key)(r)))
}
