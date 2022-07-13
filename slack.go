package main

import (
	"github.com/slack-go/slack"
	"log"
	"os"
	"time"
)

func PostToSlack(status string) {
	token := os.Getenv("SLACK_TOKEN")

	if token == "" {
		log.Println("SLACK_TOKEN not set; aborting posting slack message!")
		return
	}

	channel := os.Getenv("SLACK_CHANNEL_ID")

	if channel == "" {
		log.Println("SLACK_CHANNEL_ID not set; aborting posting slack message!")
		return
	}

	api := slack.New(token)

	channelID, _, err := api.PostMessage(channel, slack.MsgOptionText(status, false))

	if err != nil {
		log.Fatal(err)
		return
	}

	timestamp := time.Now().Local().Format(time.ANSIC)

	log.Printf("Message sent to channel %s on %s", channelID, timestamp)
}
