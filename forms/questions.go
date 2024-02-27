package forms

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"trivia/models"

	"github.com/bgaudino/godino"
)

type QuestionForm struct {
	Request    *http.Request
	Errors     map[string][]error
	Model      *models.Question
	Categories []*models.Category
}

func NewQuestionForm(r *http.Request, m *models.Question) QuestionForm {
	f := QuestionForm{Request: r, Model: m, Errors: make(map[string][]error)}
	if f.Model == nil {
		f.Model = &models.Question{
			Choices: []*models.Answer{{}, {}, {}, {}},
		}
	}
	categories, err := models.GetCategories()
	if err == nil {
		f.Categories = categories
	}
	return f
}

func (f *QuestionForm) AddEmptyQuestions() {
	for len(f.Model.Choices) < 4 {
		f.Model.Choices = append(f.Model.Choices, &models.Answer{})
	}
}

func (f *QuestionForm) IsValid() bool {
	f.Request.ParseForm()
	q := strings.TrimSpace(f.Request.Form.Get("question"))
	if q == "" {
		e := f.Errors["question"]
		e = append(e, fmt.Errorf("this field is required"))
		f.Errors["question"] = e
	}
	cat, err := strconv.Atoi(f.Request.Form.Get("category"))
	if err != nil {
		f.Errors["category"] = []error{fmt.Errorf("invalid category")}
	}
	cIsValid := false
	var category models.Category
	for _, c := range f.Categories {
		if cat == c.Id {
			category = *c
			cIsValid = true
			break
		}
	}
	if !cIsValid {
		f.Errors["category"] = []error{fmt.Errorf("invalid category")}
	} else {
		f.Model.Categories = append(f.Model.Categories, &category)
	}
	d := f.Request.Form.Get("difficulty")
	if d != "easy" && d != "medium" && d != "hard" {
		e := f.Errors["difficulty"]
		e = append(e, fmt.Errorf("difficulty must be easy, medium, or hard"))
		f.Errors["difficulty"] = e
	} else {
		f.Model.Difficulty = d
	}
	f.Model.Text = q
	correct := f.Request.Form["correct"]
	correctIndexes := godino.NewSet[int]()
	for _, c := range correct {
		i, err := strconv.Atoi(c)
		if err == nil {
			correctIndexes.Add(i)
		}
	}
	correctCount := 0
	ch := []*models.Answer{}
	for idx, c := range f.Request.Form["choices"] {
		c = strings.TrimSpace(c)
		if c != "" {
			choice := &models.Answer{Text: c}
			if correctIndexes.Has(idx) {
				choice.IsCorrect = true
				correctCount++
			}
			ch = append(ch, choice)
		}
	}
	f.Model.Choices = ch
	if len(f.Model.Choices) == 0 {
		e := f.Errors["choices"]
		e = append(e, fmt.Errorf("at least one choice is required"))
		f.Errors["choices"] = e
	} else if correctCount != 1 {
		e := f.Errors["choices"]
		e = append(e, fmt.Errorf("exactly one correct choice is required"))
		f.Errors["choices"] = e
	}
	if len(f.Errors) > 0 {
		f.AddEmptyQuestions()
		return false
	}
	return true
}

func (f *QuestionForm) Save() error {
	return f.Model.Save(nil)
}

func (f *QuestionForm) Process() {
	if f.IsValid() {
		err := f.Save()
		if err != nil {
			f.Errors["_nonFieldErrors"] = []error{err}
			f.AddEmptyQuestions()
		}
	}
}
