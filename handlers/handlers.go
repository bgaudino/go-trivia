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
	question := query["question"]
	answer := query["answer"]
	if len(question) != 1 && len(answer) != 1 {
		http.Error(w, "Must provide question and answer", http.StatusBadRequest)
		return
	}
	questionId, err := strconv.Atoi(question[0])
	if err != nil {
		http.Error(w, "Invalid Question", http.StatusBadRequest)
		return
	}
	answerId, err := strconv.Atoi(answer[0])
	if err != nil {
		http.Error(w, "Invalid Answer", http.StatusBadRequest)
		return
	}
	q, err := models.GetQuestion(questionId)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	if q == nil {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}
	templates.ExecuteTemplate(w, "_answer.html", QuestionContext{Question: q, Answer: answerId})
}
