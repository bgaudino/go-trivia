package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"trivia/db"
	"trivia/handlers"
)

func main() {
	handlers.Templates = handlers.GetTemplates()

	var err error
	db.Pool, err = db.GetPool()
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		log.Fatal("Unable to connect to database")
	}
	defer db.Pool.Close()

	http.HandleFunc("/", handlers.QuestionsHandler)
	http.HandleFunc("/api/answer/", handlers.AnswerHandler)
	http.HandleFunc("/admin/questions/add/", handlers.QuestionFormHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
