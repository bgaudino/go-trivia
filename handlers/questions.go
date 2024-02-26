package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"trivia/forms"
	"trivia/models"
)

var Templates *template.Template

func GetTemplates() *template.Template {
	dir := os.Getenv("TEMPLATES_DIR")
	if dir == "" {
		dir = "templates/"
	}
	t := template.New("main.tpl").Funcs(template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
	})
	template.Must(t.ParseGlob(dir + "*.html"))
	template.Must(t.ParseGlob(dir + "partials/*.html"))
	return t
}

func OptionsHandler(w http.ResponseWriter, r *http.Request) {
	categories, err := models.GetCategories()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
	}
	Templates.ExecuteTemplate(w, "options.html", categories)
}

func PlayHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	categoryParam := query.Get("category")
	categoryId, _ := strconv.Atoi(categoryParam)
	questions, err := models.GetQuestions(10, &models.QuestionFilters{Category: categoryId})
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
	}
	Templates.ExecuteTemplate(w, "play.html", questions)
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
		fmt.Fprintln(os.Stderr, err.Error())
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	if question == nil {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}
	if question.Answer.Id == answerId {
		w.Header().Set("HX-Trigger-After-Swap", "correct")
	} else {
		w.Header().Set("HX-Trigger-After-Swap", "incorrect")
	}
	Templates.ExecuteTemplate(w, "_answer.html", QuestionContext{Question: question, Answer: answerId})
}

func questionFormHandler(w http.ResponseWriter, r *http.Request) {
	form := forms.NewQuestionForm(r, nil)
	if r.Method == "POST" {
		form.Process()
		if len(form.Errors) > 0 {
			Templates.ExecuteTemplate(w, "question_form.html", form)
		} else {
			http.Redirect(w, r, "/", http.StatusPermanentRedirect)
		}
	} else {
		Templates.ExecuteTemplate(w, "question_form.html", form)
	}
}

var QuestionFormHandler = loginRequired(questionFormHandler)
