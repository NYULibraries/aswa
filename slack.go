package main

import (
	"github.com/slack-go/slack"
	"log"
	"os"
	"time"
)

type postMessageClient interface {
	PostMessage(channel string, options ...slack.MsgOption) (string, string, error)
}

func PostToSlack(status string) {

	token := os.Getenv("SLACK_TOKEN")

	if token == "" {
		log.Println("SLACK_TOKEN not set; aborting posting slack message!")
		return
	}

	api := slack.New(token)

	PostToSlackWithClient(status, api)

}

func PostToSlackWithClient(status string, client postMessageClient) {

	channel := os.Getenv("SLACK_CHANNEL_ID")

	if channel == "" {
		log.Println("SLACK_CHANNEL _ID not set; aborting posting slack message!")
		return
	}

	channelID, _, err := client.PostMessage(channel, slack.MsgOptionText(status, false))

	if err != nil {
		log.Fatal(err)
		return
	}

	timestamp := time.Now().Local().Format(time.ANSIC)

	log.Printf("Message sent to channel %s on %s", channelID, timestamp)

}
