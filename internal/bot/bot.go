package bot

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

type DiscordBot struct {
	token            string
	currentChannelId string
	currentSession   *discordgo.Session
}

func NewBot(token string, channelId string) *DiscordBot {
	bot := new(DiscordBot)
	bot.token = token
	bot.currentChannelId = channelId

	return bot
}

func (b *DiscordBot) Connect() error {
	var err error
	b.currentSession, err = discordgo.New("Bot " + b.token)
	//b.currentSession.AddHandler(b.messageCreate)
	return err
}

func (b DiscordBot) messageCreate(m *discordgo.MessageCreate) {
	fmt.Println("Message received: ", m.Content)
	if m.Author.ID == b.currentSession.State.User.ID {
		fmt.Println("Ignoring bot message")
		return
	}

	if len(m.Content) == 0 {
		fmt.Println("Ignoring blank message")
	}

	if len(m.Content) > 0 && m.Content[0] == '!' {
		b.handleCommand(m)
	}
}

func (b DiscordBot) handleCommand(m *discordgo.MessageCreate) {
	switch m.Content {
	case "!ping":
		b.SendMessage("pong")
	}
}

func (b DiscordBot) SendMessage(message string) {
	_, err := b.currentSession.ChannelMessageSend(b.currentChannelId, message)
	if err != nil {
		fmt.Printf("Error sending message: %v\n", err)
	}
}

func (b DiscordBot) ReadMessagesAllFromChannel() {
	messages, err := b.currentSession.ChannelMessages(b.currentChannelId, 100, "", "", "")
	if err != nil {
		fmt.Println(err)
	}
	for _, message := range messages {
		fmt.Printf("Message: %s \n", message.Content)
	}
}

func (b DiscordBot) ReadMessageInLastWeek(session *discordgo.Session, channelId string) {
	now := time.Now()
	oneWeekAgo := now.AddDate(0, 0, -7)

	messages, err := session.ChannelMessages(channelId, 100, "", "", "")
	if err != nil {
		fmt.Println(err)
	}

	var filteredMessages []*discordgo.Message
	for _, message := range messages {
		createdAt := message.Timestamp

		if createdAt.After(oneWeekAgo) {
			filteredMessages = append(filteredMessages, message)
		}
	}
}

func (b DiscordBot) Close() {
	b.currentSession.Close()
	fmt.Println("Bot session closed")
}
