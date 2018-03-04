package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/nlopes/slack"
)

type game struct {
	questions       []Question
	questionRunning bool
	players         map[string]int
	msgCh           chan string
	api             *slack.Client
}

// Question contains the information for a question
type Question struct {
	ID         int
	Question   string
	Value      int
	Answer     string
	isAnswered bool
}

var currentGame *game

// InitTrivia readies the trivia game for playing
func InitTrivia(api *slack.Client, ch chan string) {
	for {
		select {
		case msg := <-ch:
			handleMessage(api, msg)
		}
	}
}

func handleMessage(api *slack.Client, msg string) {
	if currentGame == nil { // handle message when game is not running
		handleMessageWhenGameIsNotRunning(api, msg)
	} else { // handle message when game is running
		handleMessageWhenGameIsRunning(api, msg)
	}
}

func handleMessageWhenGameIsNotRunning(api *slack.Client, msg string) {
	switch msg {
	case "start":
		startTrivia(api)
	}
}

func handleMessageWhenGameIsRunning(api *slack.Client, msg string) {
	if currentGame.questionRunning {
		currentGame.msgCh <- msg
	}
}

func startTrivia(api *slack.Client) {
	msgCh := make(chan string, 10)
	currentGame = &game{questions: getQuestions(), msgCh: msgCh, players: make(map[string]int), api: api}
	go currentGame.start()
}

func (game *game) start() {
	// Introduce the game
	PostMessage(game.api, "Starting Jeopardy")
	// https://stackoverflow.com/questions/37891280/goroutine-time-sleep-or-time-after

	// iterate over each game
	for _, currQuestion := range game.questions {
		PostMessage(game.api, "Next question in 3, 2, 1...")
		time.Sleep(time.Second * 3)
		currQuestion.read(game.api)
		game.questionRunning = true

		// accept guesses in the allowed amount of time
	GuessTime:
		for {
			select {
			case guess := <-game.msgCh:
				// PostMessage(game.api, guess)
				if currQuestion.guess(game.api, guess) {
					PostMessage(game.api, "Correct")
					// add points to player
					break GuessTime
				} else {
					PostMessage(game.api, "Incorrect")
					// deduct points from player
					// exclude them from answering this question anymore
				}
			case <-time.After(time.Second * 10):
				PostMessage(game.api, "none of you got it right lmfoa")
				break GuessTime
			}
		}

		game.questionRunning = false
		time.Sleep(time.Second * 2)
	}
}

// Read prints out the question, and readies the Question for answering
func (question *Question) read(api *slack.Client) {
	PostMessage(api, question.Question)
}

// Guess checks if a propsed answer is correct
func (question *Question) guess(api *slack.Client, answer string) bool {
	if !question.isAnswered {
		fmt.Println("Propsing answer with value: " + answer)
		correct := answer == question.Answer // might need to do some trimming of non-alphanumerics and lowercasing the guess and answer to avoid questions with odd answers

		if correct {
			fmt.Println("Correct!")
			question.isAnswered = true
			return true
		}

		fmt.Println("Incorrect. Answer was: " + question.Answer)
		return false
	}

	return false
}

func getQuestions() []Question {
	resp, err := http.Get("http://jservice.io/api/random?count=10")
	var questions = &[]Question{}

	if err == nil {
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			bad := json.Unmarshal(contents, questions)
			if bad != nil {
				fmt.Println("AHHHH")
			}
		} else {
			fmt.Println("woops again leeel")
		}
	} else {
		fmt.Println("woops lel")
	}

	return *questions
}
