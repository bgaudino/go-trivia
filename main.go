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
	var err error
	db.Pool, err = db.GetPool()
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		log.Fatal("Unable to connect to database")
	}
	defer db.Pool.Close()

	http.HandleFunc("/", handlers.QuestionsHandler)
	http.HandleFunc("/api/answer/", handlers.AnswerHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
