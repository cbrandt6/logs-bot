package bot

import (
	"fmt"
	"logs-bot/internal/textract"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

type DiscordBot struct {
	token           string
	channelId       string
	session         *discordgo.Session
	textractWrapper *textract.TextractWrapper
}

func NewBot(token string, channelId string) *DiscordBot {
	bot := new(DiscordBot)
	bot.token = token
	bot.channelId = channelId
	bot.textractWrapper = textract.NewTextractWrapper()
	return bot
}

func (b *DiscordBot) Connect() error {
	var err error
	b.session, err = discordgo.New("Bot " + b.token)
	b.session.AddHandler(b.messageCreate)
	return err
}

func (b *DiscordBot) Open() {
	b.session.Open()
	fmt.Println("Bot session opened")
}

func (b *DiscordBot) Close() {
	b.session.Close()
	fmt.Println("Bot session closed")
}

func (b *DiscordBot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Println("Message received: ", m.Content)
	if m.Author.ID == b.session.State.User.ID {
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

func (b *DiscordBot) handleCommand(m *discordgo.MessageCreate) {
	switch m.Content {
	case "!ping":
		b.SendMessage("pong")
	}
}

func (b *DiscordBot) SendMessage(message string) {
	_, err := b.session.ChannelMessageSend(b.channelId, message)
	if err != nil {
		fmt.Printf("Error sending message: %v\n", err)
	}
}

func (b *DiscordBot) GetAllMessagesFromChannel(limit int, beforeId string, afterId string, aroundId string) ([]*discordgo.Message, error) {
	messages, err := b.session.ChannelMessages(b.channelId, limit, beforeId, afterId, aroundId)

	if err != nil {
		fmt.Println(err)
	}

	return messages, err
}

func (b *DiscordBot) readMessagesInLastWeek() ([]*discordgo.Message, error) {
	oneWeekAgo := time.Now().AddDate(0, 0, -7)
	snowFlake := createSnowFlake(oneWeekAgo.UTC().UnixMilli())

	messages, err := b.GetAllMessagesFromChannel(100, "", strconv.FormatInt(snowFlake, 10), "")

	return messages, err
}

func (b *DiscordBot) CalculateScoreboard() {
	messages, err := b.readMessagesInLastWeek()

	if err != nil {
		fmt.Println(err)
	}

	for _, message := range messages {
		fmt.Println(message.Timestamp)
	}
}

func (b *DiscordBot) ParseTextFromImage() {
	lines, err := b.textractWrapper.ParseTextLinesFromImage()

	if err != nil {
		fmt.Println(err)
	}

	for _, line := range lines {
		fmt.Println(line)
	}
}

func getTimestampFromSnowFlake(snowflake string) (time.Time, error) {
	snowflakeAsInt, err := strconv.ParseInt(snowflake, 10, 64)

	if err != nil {
		fmt.Println("Error:", err)
	}

	return time.UnixMilli(((snowflakeAsInt >> 22) + 1420070400000)).UTC(), err
}

func createSnowFlake(timestamp int64) int64 {
	const discordEpoch = 1420070400000
	return (timestamp - discordEpoch) << 22
}
