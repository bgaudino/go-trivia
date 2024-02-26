package models

import (
	"context"
	"errors"
	"fmt"
	"os"
	"trivia/db"
	"trivia/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Answer struct {
	Id        int
	Text      string
	IsCorrect bool
}

type Category struct {
	Id   int
	Name string
}

func GetCategories() ([]*Category, error) {
	categories := []*Category{}
	rows, err := db.Pool.Query(context.Background(), "SELECT id, name FROM categories ORDER BY name")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		c := Category{}
		err := rows.Scan(&c.Id, &c.Name)
		if err != nil {
			return nil, err
		}
		categories = append(categories, &c)
	}
	return categories, nil
}

type Question struct {
	Id         int
	Text       string
	Choices    []*Answer
	Answer     *Answer
	Difficulty string
	Categories []*Category
}

func (q *Question) Save(conn *pgxpool.Pool) error {
	if conn == nil {
		conn = db.Pool
	}
	ctx := context.Background()
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			rbErr := tx.Rollback(ctx)
			if rbErr != nil {
				fmt.Fprintln(os.Stderr, rbErr.Error())
			}
		}
	}()

	if q.Difficulty == "" {
		q.Difficulty = "medium"
	}
	err = tx.QueryRow(
		ctx,
		"INSERT INTO questions (text, difficulty) VALUES ($1, $2) RETURNING id",
		q.Text, q.Difficulty,
	).Scan(&q.Id)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.ConstraintName == "questions_text_key" {
				return errors.New("this question already exists")
			}
		}
		return err
	}

	for _, c := range q.Choices {
		err = tx.QueryRow(
			ctx,
			"INSERT INTO answers (text, question_id, is_correct) VALUES ($1, $2, $3) RETURNING id",
			c.Text, q.Id, c.IsCorrect,
		).Scan(&c.Id)
		if err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok {
				if pgErr.ConstraintName == "answers_text_question_id_key" {
					return errors.New("choices must be unique per question")
				}
			}
			return err
		}
	}
	return tx.Commit(ctx)
}

type QuestionFilters struct {
	Category int
}

func GetQuestions(n int, filters *QuestionFilters) (*utils.OrderedMap[int, *Question], error) {
	questions := utils.NewOrderedMap[int, *Question]()
	query := "SELECT id, text FROM questions ORDER BY RANDOM() LIMIT $1"
	params := []any{n}
	if filters.Category != 0 {
		params = append(params, filters.Category)
		query = `
			SELECT questions.id, questions.text FROM questions
			JOIN categorization ON questions.id = categorization.question_id 
			WHERE categorization.category_id = $2 
			ORDER BY RANDOM() LIMIT $1
		`
	}
	rows, err := db.Pool.Query(
		context.Background(),
		query,
		params...,
	)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var q = Question{Choices: []*Answer{}}
		rows.Scan(&q.Id, &q.Text)
		questions.Insert(q.Id, &q)
	}
	rows, err = db.Pool.Query(
		context.Background(),
		"SELECT id, text, is_correct, question_id FROM answers WHERE question_id = ANY($1) ORDER BY id",
		questions.Keys(),
	)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int
		var answer Answer
		rows.Scan(&answer.Id, &answer.Text, &answer.IsCorrect, &id)
		q := questions.Get(id)
		q.Choices = append(q.Choices, &answer)
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
			ORDER BY answers.id
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
