package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/miguelmota/tweet-patterns/client"
)

var errUsername = errors.New("username is required")

func main() {
	if len(os.Args) < 2 {
		panic(errUsername)
	}

	username := os.Args[1]
	if username == "" {
		panic(errUsername)
	}

	tp := client.NewClient(&client.Config{
		ConsumerKey:       os.Getenv("TWITTER_CONSUMER_KEY"),
		ConsumerSecret:    os.Getenv("TWITTER_CONSUMER_SECRET"),
		AccessTokenKey:    os.Getenv("TWITTER_ACCESS_TOKEN_KEY"),
		AccessTokenSecret: os.Getenv("TWITTER_ACCESS_TOKEN_SECRET"),
		Username:          username,
	})

	filename, err := tp.Save()
	if err != nil {
		panic(err)
	}

	fmt.Printf("saved %s\n", filename)
}
