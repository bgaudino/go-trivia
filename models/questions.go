package models

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
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

	for _, c := range q.Categories {
		_, err = tx.Exec(
			ctx,
			"INSERT INTO categorization (question_id, category_id) VALUES ($1, $2)",
			q.Id, c.Id,
		)
		if err != nil {
			return err
		}
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
	Category   int
	Difficulty string
	Count      int
}

func GetQuestions(filters *QuestionFilters) (*utils.OrderedMap[int, *Question], error) {
	questions := utils.NewOrderedMap[int, *Question]()
	var query strings.Builder
	params := []any{filters.Count}
	query.WriteString("SELECT questions.id, questions.text, questions.difficulty FROM questions")
	if filters.Category != 0 {
		params = append(params, filters.Category)
		query.WriteString(`
			JOIN categorization ON questions.id = categorization.question_id
			WHERE categorization.category_id = $2
		`)
	}
	if filters.Difficulty != "" {
		params = append(params, filters.Difficulty)
		l := len(params)
		if l > 2 {
			query.WriteString(" AND ")
		} else {
			query.WriteString(" WHERE ")
		}
		query.WriteString(fmt.Sprintf("difficulty = $%v", l))
	}
	query.WriteString(" ORDER BY RANDOM() LIMIT $1")
	rows, err := db.Pool.Query(
		context.Background(),
		query.String(),
		params...,
	)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var q = Question{Choices: []*Answer{}}
		rows.Scan(&q.Id, &q.Text, &q.Difficulty)
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
