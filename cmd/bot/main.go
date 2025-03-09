package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

//go:embed praises.json
var praisesFile embed.FS

//go:embed conspiracies.json
var conspiraciesFile embed.FS

var (
	praises      []string
	conspiracies []string
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	botToken := os.Getenv("DISCORD_BOT_TOKEN")
	channelID := os.Getenv("DISCORD_CHANNEL_ID")
	if botToken == "" || channelID == "" {
		fmt.Println("Missing DISCORD_BOT_TOKEN or DISCORD_CHANNEL_ID in .env")
		return
	}

	// load messages at build time
	err = loadMessages()
	if err != nil {
		fmt.Println("Error loading embedded messages: %v", err)
		return
	}

	// initialize bot session
	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		fmt.Println("Error creating Discord session,", err)
		return
	}

	// open WebSocket connection
	dg.AddHandler(readyHandler)
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening connection,", err)
		return
	}

	// start the scheduled messages
	go sendScheduledMessages(dg, channelID)

	// wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// cleanly close down the Discord session
	dg.Close()
}

func loadMessages() error {
	file, err := praisesFile.ReadFile("praises.json")
	if err != nil {
		return err
	}

	err = json.Unmarshal(file, &praises)
	if err != nil {
		return err
	}

	file, err = conspiraciesFile.ReadFile("conspiracies.json")
	if err != nil {
		return err
	}

	err = json.Unmarshal(file, &conspiracies)
	if err != nil {
		return err
	}

	return nil
}

func readyHandler(s *discordgo.Session, r *discordgo.Ready) {
	fmt.Println("Ready to rumble!")
}

func sendScheduledMessages(s *discordgo.Session, channelID string) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		message := "Good"
		_, err := s.ChannelMessageSend(channelID, message)
		if err != nil {
			fmt.Printf("Error sending message: %v\n", err)
		}
	}
}
