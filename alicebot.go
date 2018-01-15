package main

import (
	"encoding/json"
	"log"
	"os"
	"fmt"
	"github.com/nlopes/slack"
	"strings"
)

type Configuration struct {
	AdminUserId string
	DefaultChannelId string
	Token string
	LogFile string
}

var configuration = Configuration{}

// Global flags
var rest_flag = false
var break_loop_flag = false

func main() {
	ReadConfiguration("config.json")

	api := slack.New(configuration.Token)
	api.SetDebug(true)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

Loop:
	for {
		if break_loop_flag {
			break Loop
		}
	
		select {
		case msg := <-rtm.IncomingEvents:
			fmt.Print("Event Received: ")
			switch ev := msg.Data.(type) {
			case *slack.HelloEvent:
				// Ignore hello

			case *slack.ConnectedEvent:
				fmt.Println("Infos:", ev.Info)
				fmt.Println("Connection counter:", ev.ConnectionCount)
				// Replace #general with your Channel ID
				rtm.SendMessage(rtm.NewOutgoingMessage("Hello world", "#general"))

			case *slack.MessageEvent:
				fmt.Printf("Message: %v\n", ev.Text)
				HandleMessage(api, ev)

			case *slack.PresenceChangeEvent:
				fmt.Printf("Presence Change: %v\n", ev)

			case *slack.LatencyReport:
				fmt.Printf("Current latency: %v\n", ev.Value)

			case *slack.RTMError:
				fmt.Printf("Error: %s\n", ev.Error())

			case *slack.InvalidAuthEvent:
				fmt.Printf("Invalid credentials")
				break Loop

			default:

				// Ignore other events..
				// fmt.Printf("Unexpected: %v\n", msg.Data)
			}
		}
	}
}

// Readies config values
func ReadConfiguration(fileName string) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal("Error opening configuration file")
	}
	decoder := json.NewDecoder(file)
	configuration = Configuration{}
	err = decoder.Decode(&configuration)
	if(err != nil) {
		log.Fatal("Error parsing configuration file")
	}
}

// Message handler
// Handlers: alice, pokemon
func HandleMessage(api *slack.Client, ev *slack.MessageEvent) {
	words := strings.Split(ev.Text, " ")
	
	if len(words) <= 0 {
		return
	} else {
		// Check what type of command I'm going to handle
		switch words[0] {
		
		// Alice (admin) events, admin only, ignore resting
		case "alice":
			if strings.EqualFold(ev.User, configuration.AdminUserId){
				HandleAliceMessage(api, words)
			}
		
		// Pokemon events, all users
		case "pokemon":
			if !rest_flag {
				// HandlePokemonMessage(api, words)
			}
		}
	}
}

// Super User message handler
func HandleAliceMessage(api *slack.Client, words []string) {

	if len(words) <= 1 {
		PostMessage(api, "Did you need something?")
	} else {
	
		// Check what type of alice command I'm going to handle
		switch words[1] {
		
			// Awake
			case "awake":
				if rest_flag {
					rest_flag = false
					PostMessage(api, "I'm awake! I'm awake!")
				}
			
			// Morning
			case "morning":
				PostMessage(api, "Good morning!")
			
			// Rest
			case "rest":
				rest_flag = true
				PostMessage(api, "....zzzZZZZzzzZZzz...")
				
			// Sleep
			case "sleep":
				PostMessage(api, "Bye bye!")
				break_loop_flag = true
			
		}
	}
}

func HandlePokemonMessage(api *slack.Client, words []string) {
	PostMessage(api, "POKEMON!?!?!?!")
}

// Posts the message to the chat (channel, name, icon are hardcoded)
func PostMessage(api *slack.Client, message string) {
	params := slack.PostMessageParameters{}
	attachment := slack.Attachment{
		// Uncomment the following part to send a field too
		/*
			Fields: []slack.AttachmentField{
				slack.AttachmentField{
					Title: "a",
					Value: "no",
				},
			},
		*/
	}
	params.Attachments = []slack.Attachment{attachment}
	params.Username = "alicebot"
	params.AsUser = false
	params.IconURL = "http://www.outback-australia-travel-secrets.com/image-files/australian-spiders-redback2.jpg"

	channelID, timestamp, err := api.PostMessage(configuration.DefaultChannelId, message, params)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	
	fmt.Printf("Message successfully sent to channel %s at %s:\n%s", channelID, timestamp, message)

}
