package bot

import (
	"fmt"
	"io"
	"logs-bot/internal/textract"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type DiscordBot struct {
	token           string
	channelId       string
	session         *discordgo.Session
	textractWrapper *textract.TextractWrapper
}

type UserStats struct {
	userId      string
	username    string
	totalWeight int
	totalSets   int
}

type ImageStats struct {
	totalWeight int
	totalSets   int
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
	oneWeekAgo = time.Date(oneWeekAgo.Year(), oneWeekAgo.Month(), oneWeekAgo.Day(), 0, 0, 0, 0, oneWeekAgo.Location())
	snowFlake := createSnowFlake(oneWeekAgo.UTC().UnixMilli())

	messages, err := b.GetAllMessagesFromChannel(100, "", strconv.FormatInt(snowFlake, 10), "")

	return messages, err
}

func (b *DiscordBot) CalculateScoreboard() {
	messages, err := b.readMessagesInLastWeek()
	if err != nil {
		fmt.Println(err)
	}
	userStats := make(map[string]UserStats)

	if len(messages) == 0 {
		fmt.Println("No recent messages!")
		return
	}

	for _, message := range messages {
		if len(message.Attachments) == 0 {
			continue
		}
		stats, err := b.processAttachments(message.Attachments)
		if err != nil {
			fmt.Println(err)
		}

		// Checking if we've already seen this user
		userId := message.Author.ID
		userStat, exists := userStats[userId]

		if exists {
			userStat.totalWeight += stats.totalWeight
			userStat.totalSets += stats.totalSets
			userStats[userId] = userStat
		} else {
			newUserStat := UserStats{
				userId:      message.Author.ID,
				username:    message.Author.Username,
				totalWeight: stats.totalWeight,
				totalSets:   stats.totalSets}

			userStats[userId] = newUserStat
		}
	}

	for _, val := range userStats {
		fmt.Printf("Stats for %s\n", val.username)
		fmt.Printf("Total weight: %d, Total sets: %d \n", val.totalWeight, val.totalSets)
	}
}

func (b *DiscordBot) processAttachments(attachments []*discordgo.MessageAttachment) (ImageStats, error) {
	var stats ImageStats
	weightRegex := regexp.MustCompile(`(\d{1,3}(,\d{3})*\s?lb)`)
	setRegex := regexp.MustCompile(`(\d+\s?sets)`)

	for _, attachment := range attachments {
		if isImage(attachment.Filename) {
			// Using id and filename to ensure the files are unique
			filename := attachment.ID + attachment.Filename
			err := b.downloadImage(filename, attachment.URL)
			if err != nil {
				fmt.Println(err)
			}

			// Get text from image
			lines, err := b.ParseTextFromImage(filename)
			if err != nil {
				fmt.Println(err)
			}

			for _, line := range lines {
				// At time of writing the lbs will be before sets
				matches := weightRegex.FindAllString(line, -1)
				if len(matches) > 0 {
					weight, err := parseIntFromString(matches[0])
					if err != nil {
						fmt.Println(err)
					}

					stats.totalWeight += weight
					continue
				}
				setMatches := setRegex.FindAllString(line, -1)
				if len(setMatches) > 0 {
					numSets, err := parseIntFromString(setMatches[0])
					if err != nil {
						fmt.Println(err)
					}

					stats.totalSets += numSets
					// We don't care about anything after the sets
					break
				}
			}
			os.Remove(filename)
		}
	}
	return stats, nil
}

func (b *DiscordBot) ParseTextFromImage(filename string) ([]string, error) {
	lines, err := b.textractWrapper.ParseTextLinesFromImage(textract.ImageDir + "/" + filename)
	if err != nil {
		fmt.Println(err)
	}

	return lines, err
}

func (b *DiscordBot) downloadImage(filename string, url string) error {
	file, err := os.Create(textract.ImageDir + "/" + filename)
	if err != nil {
		return err
	}

	response, err := http.Get(url)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func isImage(filename string) bool {
	switch filename[len(filename)-4:] {
	case ".png", ".jpg", ".jpeg":
		return true
	default:
		return false
	}
}

func createSnowFlake(timestamp int64) int64 {
	const discordEpoch = 1420070400000
	return (timestamp - discordEpoch) << 22
}

func parseIntFromString(str string) (int, error) {
	cleanStr := strings.Replace(str, "lb", "", -1)
	cleanStr = strings.Replace(cleanStr, "sets", "", -1)
	cleanStr = strings.Replace(cleanStr, ",", "", -1)
	cleanStr = strings.Replace(cleanStr, " ", "", -1)
	str64, err := strconv.ParseInt(cleanStr, 10, 0)

	if err != nil {
		fmt.Println(err)
	}
	intStr := int(str64)

	return intStr, err
}
