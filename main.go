package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"

	language "cloud.google.com/go/language/apiv1"
	"github.com/bwmarrin/discordgo"
	languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"
)

// Bot parameters
var (
	GuildID  = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	BotToken = flag.String("token", "", "Bot access token")
	BotName  = flag.String("name", "Ursbot", "Bot name")
)

var s *discordgo.Session
var stopbot = make(chan struct{})

func init() {
	flag.Parse()
	if *BotToken == "" {
		fpath := os.Getenv("BOT_TOKEN")
		f, err := os.Open(fpath)
		if err != nil {
			log.Fatal(err)
		}
		tokenByte, err := io.ReadAll(f)
		if err != nil {
			log.Fatal(err)
		}
		*BotToken = strings.Trim(string(tokenByte), "\n")
	}
	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

func checkEmo(s string) float64 {
	ctx := context.Background()

	// Creates a client.
	client, err := language.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Detects the sentiment of the text.
	sentiment, err := client.AnalyzeSentiment(ctx, &languagepb.AnalyzeSentimentRequest{
		Document: &languagepb.Document{
			Source: &languagepb.Document_Content{
				Content: s,
			},
			Type: languagepb.Document_PLAIN_TEXT,
		},
		EncodingType: languagepb.EncodingType_UTF8,
	})
	if err != nil {
		log.Fatalf("Failed to analyze text: %v", err)
	}

	fmt.Printf("Text: %v\n", s)
	var sum float64
	sum = float64(sentiment.DocumentSentiment.Score)
	if sum >= 0 {
		fmt.Println("Sentiment: positive")
	} else {
		fmt.Println("Sentiment: negative")
	}
	return sum
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		channelID := m.ChannelID
		authorID := m.Author.ID
		if !strings.Contains(m.Content, "募") {
			return
		}

		log.Println("募集を開始")

		s.AddHandlerOnce(func(s *discordgo.Session, m *discordgo.MessageCreate) {
			if channelID != m.ChannelID {
				return
			}
			if m.Author.ID == s.State.User.ID && authorID == m.Author.ID {
				return
			}

			if checkEmo(m.Content) >= 0 {
				_, err := s.ChannelMessageSend(m.ChannelID, "よし、じゃあいこう!")
				if err != nil {
					log.Println("Error sending message: ", err)
				}
				return
			} else {
				_, err := s.ChannelMessageSend(m.ChannelID, "https://pics.prcm.jp/01011214/49849807/jpeg/49849807.jpeg")
				if err != nil {
					log.Println("Error sending message: ", err)
				}
			}
		})
		log.Println("募集を終了")
	})

	err := s.Open()
	defer s.Close()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	log.Println("Gracefully shutting down.")
}
