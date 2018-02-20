package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/nlopes/slack"
)

// Question contains the information for a question
type Question struct {
	ID         int
	Question   string
	Value      int
	Answer     string
	isAnswered bool
}

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
	switch msg {
	case "start":
		PostMessage(api, "Starting Jeopardy")
		playTrivia(api, []string{"Eric", "Andrew"})
	}
}

// layTrivia begins a game of jeopardy
func playTrivia(api *slack.Client, players []string) {
	if len(players) == 0 {
		// fmt.Println("No players entered.")
		PostMessage(api, "No players entered.")
		return
	}

	// Introduce the game
	// fmt.Println("Welcome to Jeopardy. I'll be your host, Alice Trebek.")
	PostMessage(api, "Welcome to Jeopardy. I'll be your host, Alice Trebek.")
	time.Sleep(2000 * time.Millisecond)
	introducePlayers(api, players)
}

func introducePlayers(api *slack.Client, players []string) {
	firstName := true
	var introduction bytes.Buffer
	introduction.WriteString("Our contestants today are: ")
	// fmt.Print("Our contestants today are: ")

	for _, player := range players {
		if !firstName {
			// fmt.Print(", ")
			introduction.WriteString(", ")
		}
		// fmt.Print(player)
		introduction.WriteString(player)
		firstName = false
	}

	PostMessage(api, introduction.String())
}

// Read prints out the question, and readies the Question for answering
func (question *Question) read() {
	fmt.Println(question.Question)
}

// Guess checks if a propsed answer is correct
func (question *Question) guess(answer string) {
	if !question.isAnswered {
		fmt.Println("Propsing answer with value: " + answer)
		correct := answer == question.Answer // might need to do some trimming of non-alphanumerics and lowercasing the guess and answer to avoid questions with odd answers

		if correct {
			fmt.Println("Correct!")
			question.isAnswered = true
		} else {
			fmt.Println("Incorrect. Answer was: " + question.Answer)
		}
	}
}

func getQuestion() *Question {
	resp, err := http.Get("http://jservice.io/api/random?count=1")
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

	(*questions)[0].isAnswered = false
	return &(*questions)[0]
}
