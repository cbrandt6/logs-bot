package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/textract"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var textractSession *textract.Textract
var token string

func init() {

	err := godotenv.Load()
	if err != nil {
		panic("Unable to load env file")
	}

	textractSession = textract.New(session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})))

	token = os.Getenv("discord_token")
}

func parseTextFromImage() {
	file, err := os.ReadFile("sample.png")
	if err != nil {
		panic(err)
	}

	resp, err := textractSession.DetectDocumentText(&textract.DetectDocumentTextInput{
		Document: &textract.Document{
			Bytes: file,
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(resp)

	for i := 1; i < len(resp.Blocks); i++ {
		if *resp.Blocks[i].BlockType == "WORD" {
			fmt.Println(*resp.Blocks[i].Text)
		}
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Println("Message received: ", m.Content)
	if m.Author.ID == s.State.User.ID {
		fmt.Println("Ignoring bot message")
		return
	}

	if len(m.Content) == 0 {
		fmt.Println("Ignoring blank message")
	}

	if len(m.Content) > 0 && m.Content[0] == '!' {
		handleCommand(s, m)
	}
}

func handleCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	switch m.Content {
	case "!ping":
		sendMessage(s, m.ChannelID, "pong")
	}
}

func sendMessage(s *discordgo.Session, channelID, message string) {
	_, err := s.ChannelMessageSend(channelID, message)
	if err != nil {
		fmt.Printf("Error sending message: %v\n", err)
	}
}

func main() {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session", err)
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}
