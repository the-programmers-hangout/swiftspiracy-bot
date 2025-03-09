package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
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
