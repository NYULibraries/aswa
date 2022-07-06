package main

import (
	"github.com/slack-go/slack"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func parseUnixTimestamp(ms string) (time.Time, error) {
	str := strings.Split(ms, ".")[0]
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		panic(err)
	}

	tm := time.Unix(i, 0)

	return tm, nil
}

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

	channelID, timestamp, err := api.PostMessage(channel, slack.MsgOptionText(status, false))

	if err != nil {
		log.Fatal(err)
		return
	}

	parsed_time, _ := parseUnixTimestamp(timestamp)

	log.Printf("Message sent to channel %s at %s", channelID, parsed_time)
}
