package main

import (
	"fmt"
	"logs-bot/internal/bot"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

var token string
var channelID string

func init() {

	err := godotenv.Load()
	if err != nil {
		panic("Unable to load env file")
	}

	token = os.Getenv("discord_token")
	channelID = os.Getenv("dev_channel_id")
}

func main() {
	bot := bot.NewBot(token, channelID)
	err := bot.Connect()
	if err != nil {
		fmt.Println("Error connecting", err)
		return
	}
	bot.ReadMessagesAllFromChannel()

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGTERM)
	<-sc

	bot.Close()
}
