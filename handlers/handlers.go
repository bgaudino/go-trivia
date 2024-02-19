package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"trivia/models"
)

var templates = template.Must(template.ParseFiles("templates/index.html", "templates/partials/_answer.html"))

func QuestionsHandler(w http.ResponseWriter, r *http.Request) {
	questions, err := models.GetAllQuestions()
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
	}
	templates.ExecuteTemplate(w, "index.html", &questions)
}

type QuestionContext struct {
	Question *models.Question
	Answer   int
}

func AnswerHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	questionParam := query["question"]
	answerParam := query["answer"]
	if len(questionParam) == 0 || len(answerParam) == 0 {
		http.Error(w, "Must provide question and answer", http.StatusBadRequest)
		return
	}
	questionId, err := strconv.Atoi(questionParam[0])
	if err != nil {
		http.Error(w, "Invalid Question", http.StatusBadRequest)
		return
	}
	answerId, err := strconv.Atoi(answerParam[0])
	if err != nil {
		http.Error(w, "Invalid Answer", http.StatusBadRequest)
		return
	}
	question, err := models.GetQuestion(questionId)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	if question == nil {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}
	templates.ExecuteTemplate(w, "_answer.html", QuestionContext{Question: question, Answer: answerId})
}
