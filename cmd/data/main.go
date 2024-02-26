package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"math/big"
	"net/http"
	"time"
	"trivia/db"
	"trivia/models"
)

type TriviaResult struct {
	Type             string   `json:"type"`
	Difficulty       string   `json:"difficulty"`
	Category         string   `json:"category"`
	Question         string   `json:"question"`
	CorrectAnswer    string   `json:"correct_answer"`
	IncorrectAnswers []string `json:"incorrect_answers"`
}
type TriviaResponse struct {
	ResponseCode int            `json:"response_code"`
	Results      []TriviaResult `json:"results"`
}

type TokenResponse struct {
	ResponseCode int    `json:"response_code"`
	Token        string `json:"token"`
}

func getChoices(result TriviaResult) []*models.Answer {
	choices := []*models.Answer{}
	correctIndex, _ := randomInt(len(result.IncorrectAnswers))
	j := 0
	for i := 0; i <= len(result.IncorrectAnswers); i++ {
		if i == correctIndex {
			choices = append(choices, &models.Answer{
				Text:      html.UnescapeString(result.CorrectAnswer),
				IsCorrect: true,
			})
		} else {
			choices = append(choices, &models.Answer{Text: html.UnescapeString(result.IncorrectAnswers[j])})
			j++
		}
	}
	return choices
}

func randomInt(n int) (int, error) {
	max := big.NewInt(int64(n))
	randomNumber, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}
	return int(randomNumber.Int64()), nil
}

func getQuestions(token string) (*TriviaResponse, error) {
	res, err := http.Get("https://opentdb.com/api.php?amount=50&token=" + token)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var q TriviaResponse
	err = json.Unmarshal(body, &q)
	if err != nil {
		return nil, err
	}
	return &q, nil
}

func getToken() (*TokenResponse, error) {
	res, err := http.Get("https://opentdb.com/api_token.php?command=request")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var token TokenResponse
	json.Unmarshal(body, &token)
	return &token, nil
}

func main() {
	fail := func(err error) {
		log.Fatal(err.Error())
	}
	pool, err := db.GetPool()
	if err != nil {
		fail(err)
	}
	defer pool.Close()
	categories := map[string]int{}
	rows, err := pool.Query(context.Background(), "SELECT id, name FROM categories")
	if err != nil {
		fail(err)
	}
	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		categories[name] = id
	}
	res := &TriviaResponse{}
	token, err := getToken()
	if err != nil {
		fail(err)
	}
	fmt.Print("Saving questions [")
	questionsSaved := 0
	for res.ResponseCode == 0 {
		res, err = getQuestions(token.Token)
		if err != nil {
			fail(err)
		}
		for _, r := range res.Results {
			text := html.UnescapeString(r.Question)
			rows := pool.QueryRow(context.Background(), "SELECT id FROM questions WHERE text = $1", text)
			var questionId int
			err = rows.Scan(&questionId)
			if err == nil {
				continue
			}
			q := models.Question{Text: text, Difficulty: r.Difficulty}
			q.Choices = getChoices(r)
			err := q.Save(pool)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			categoryId, ok := categories[r.Category]
			if !ok {
				row := pool.QueryRow(
					context.Background(),
					"INSERT INTO categories (name) VALUES ($1) RETURNING id",
					r.Category,
				)
				err := row.Scan(&categoryId)
				if err == nil {
					categories[r.Category] = categoryId
				}
			}
			_, err = pool.Exec(
				context.Background(),
				"INSERT INTO categorization (question_id, category_id) VALUES ($1, $2)",
				&q.Id, &categoryId,
			)
			if err != nil {
				fail(err)
			}
			questionsSaved++
		}
		fmt.Printf("%v..", questionsSaved)
		time.Sleep(5 * time.Second) // Each IP can only access the API once every 5 seconds.
	}
	fmt.Print("]\n")
}
