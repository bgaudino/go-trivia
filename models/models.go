package models

import (
	"context"
	"trivia/db"
)

type Answer struct {
	Id        int
	Text      string
	IsCorrect bool
}

type Question struct {
	Id      int
	Text    string
	Choices []*Answer
	Answer  *Answer
}

func GetAllQuestions() (map[int]Question, error) {
	questions := map[int]Question{}
	row, err := db.Pool.Query(
		context.Background(),
		`
			SELECT questions.id, questions.text, answers.id, answers.text, answers.is_correct 
			FROM questions 
			JOIN answers ON answers.question_id = questions.id
		`,
	)
	if err != nil {
		return nil, err
	}
	for row.Next() {
		var id int
		var question string
		var answer Answer
		row.Scan(&id, &question, &answer.Id, &answer.Text, &answer.IsCorrect)
		if q, ok := questions[id]; ok {
			q.Choices = append(q.Choices, &answer)
			questions[id] = q
		} else {
			questions[id] = Question{
				Text:    question,
				Choices: []*Answer{&answer},
			}
		}
		if answer.IsCorrect {
			q := questions[id]
			q.Answer = &answer
			questions[id] = q
		}
	}
	return questions, nil
}

func GetQuestion(id int) (*Question, error) {
	var question Question
	row, err := db.Pool.Query(
		context.Background(),
		`
			SELECT questions.id, questions.text, answers.id, answers.text, answers.is_correct 
			FROM questions 
			JOIN answers ON answers.question_id = questions.id
			WHERE questions.id = $1
		`, id,
	)
	if err != nil {
		return nil, err
	}
	for row.Next() {
		var answer Answer
		row.Scan(&question.Id, &question.Text, &answer.Id, &answer.Text, &answer.IsCorrect)
		question.Choices = append(question.Choices, &answer)
		if answer.IsCorrect {
			question.Answer = &answer
		}
	}
	if len(question.Choices) == 0 {
		return nil, nil
	}
	return &question, nil
}
