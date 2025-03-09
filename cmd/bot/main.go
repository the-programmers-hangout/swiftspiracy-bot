package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// Embed JSON files at build time

//go:embed praises.json
var praisesFile embed.FS

//go:embed conspiracies.json
var conspiraciesFile embed.FS

var (
	praises      []string
	conspiracies []string

	praiseIndex     int
	conspiracyIndex int
)

const (
	// TODO: update value after debugging
	SendMessageIntervalMin = 5
	// TODO: update value after debugging
	SendMessageIntervalMax = 10
	// TODO: update value after debugging
	SendMessageUnit = time.Second
	// TODO: update value after debugging
	DeleteConspiracyDelay = 3 * time.Second
	// TODO: update value after debugging
	ConspiracyProbability = 0.4
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("[!] No .env file found, using system environment variables.")
	}

	// Retrieve token and channel ID
	botToken, channelID := os.Getenv("DISCORD_BOT_TOKEN"), os.Getenv("DISCORD_CHANNEL_ID")
	if botToken == "" || channelID == "" {
		log.Fatal("[x] Missing DISCORD_BOT_TOKEN or DISCORD_CHANNEL_ID in .env")
	}

	// Load messages at build time
	if err := loadMessages(); err != nil {
		log.Fatalf("[x] Error loading messages: %v", err)
	}

	// Initialize bot session
	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatalf("[x] Error creating Discord session: %v", err)
	}

	// Open WebSocket connection
	dg.AddHandler(readyHandler)
	if err := dg.Open(); err != nil {
		log.Fatalf("[x] Error opening connection: %v", err)
	}
	defer dg.Close()

	// Start the message scheduler
	go startScheduler(dg, channelID)

	// Graceful shutdown handling
	log.Println("[✓] Swiftspiracy Bot is now running. Press CTRL+C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("[↓] Shutting down bot gracefully...")
}

// loadMessages loads JSON files into memory at build time
func loadMessages() error {
	var err error
	if praises, err = loadJSONFromEmbed(praisesFile, "praises.json"); err != nil {
		return fmt.Errorf("failed to load praises.json: %w", err)
	}
	if conspiracies, err = loadJSONFromEmbed(conspiraciesFile, "conspiracies.json"); err != nil {
		return fmt.Errorf("failed to load conspiracies.json: %w", err)
	}
	return nil
}

// loadJSONFromEmbed reads JSON from an embedded file system
func loadJSONFromEmbed(fs embed.FS, filename string) ([]string, error) {
	data, err := fs.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var messages []string
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, err
	}
	return messages, nil
}

// readyHandler confirms the bot is online
func readyHandler(s *discordgo.Session, r *discordgo.Ready) {
	log.Println("[✓] Swiftspiracy Bot is online and ready!")
}

// startScheduler handles sending messages at intervals
func startScheduler(s *discordgo.Session, channelID string) {
	randomDuration := 0 * time.Second

	for {
		// Wait for the randomized time before sending a message
		time.Sleep(randomDuration)
		sendMessage(praises[praiseIndex%len(praises)], s, channelID)
		praiseIndex++

		// Chance to send a conspiracy theory
		if rand.Float32() < ConspiracyProbability {
			discordMessage := sendMessage(conspiracies[conspiracyIndex%len(conspiracies)], s, channelID)
			conspiracyIndex++

			go deleteMessageAfterDelay(discordMessage, s, channelID)
		}

		randomDuration = time.Duration(rand.Intn(SendMessageIntervalMin)+(SendMessageIntervalMax-SendMessageIntervalMin)) * SendMessageUnit
		log.Printf("[!] Next message in %v", randomDuration)
	}
}

// sendMessage sends a random message from a given list
func sendMessage(message string, s *discordgo.Session, channelID string) *discordgo.Message {
	discordMessage, err := s.ChannelMessageSend(channelID, message)
	if err != nil {
		log.Printf("[x] Error sending message: %v\n", err)
		return nil
	}
	return discordMessage
}

// deleteMessageAfterDelay deletes the message after a given delay
func deleteMessageAfterDelay(discordMessage *discordgo.Message, s *discordgo.Session, channelID string) {
	if discordMessage == nil {
		return
	}

	time.Sleep(DeleteConspiracyDelay)
	err := s.ChannelMessageDelete(channelID, discordMessage.ID)
	if err != nil {
		log.Printf("[!] Error deleting conspiracy: %v\n", err)
	} else {
		log.Println("[↓] Conspiracy deleted!")
	}
}
