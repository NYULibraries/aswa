package main

import (
	"github.com/slack-go/slack"
	"log"
	"os"
)

func PostToSlack(status string) {
	token := os.Getenv("SLACK_TOKEN")

	if token == "" {
		log.Println("SLACK_TOKEN not set; aborting!")
		return
	}

	channel := os.Getenv("CHANNEL_ID")

	if channel == "" {
		log.Println("CHANNEL_ID not set; aborting!")
		return
	}

	api := slack.New(token)

	channelID, timestamp, err := api.PostMessage(channel, slack.MsgOptionText(status, false))

	if err != nil {
		log.Fatal(err)
		return
	}

	log.Printf("Message sent to channel %s at %s", channelID, timestamp)
}
