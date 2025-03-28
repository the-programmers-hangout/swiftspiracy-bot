package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
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

var (
	SendMessageIntervalMin int
	SendMessageIntervalMax int
	SendMessageUnit        time.Duration
	DeleteConspiracyDelay  time.Duration
	ConspiracyProbability  float64
)

func loadEnvConfig() error {
	var err error

	// Message intervals
	SendMessageIntervalMin, err = strconv.Atoi(os.Getenv("SEND_MESSAGE_INTERVAL_MIN"))
	if err != nil {
		return fmt.Errorf("SEND_MESSAGE_INTERVAL_MIN: %w", err)
	}
	SendMessageIntervalMax, err = strconv.Atoi(os.Getenv("SEND_MESSAGE_INTERVAL_MAX"))
	if err != nil {
		return fmt.Errorf("SEND_MESSAGE_INTERVAL_MAX: %w", err)
	}

	// Duration unit
	unit := os.Getenv("SEND_MESSAGE_UNIT")
	switch unit {
	case "second", "seconds":
		SendMessageUnit = time.Second
	case "minute", "minutes":
		SendMessageUnit = time.Minute
	case "millisecond", "milliseconds":
		SendMessageUnit = time.Millisecond
	default:
		return fmt.Errorf("unsupported SEND_MESSAGE_UNIT: %s", unit)
	}

	// Conspiracy delete delay
	DeleteConspiracyDelay, err = time.ParseDuration(os.Getenv("DELETE_CONSPIRACY_DELAY"))
	if err != nil {
		return fmt.Errorf("DELETE_CONSPIRACY_DELAY: %w", err)
	}

	// Conspiracy probability
	ConspiracyProbability, err = strconv.ParseFloat(os.Getenv("CONSPIRACY_PROBABILITY"), 64)
	if err != nil {
		return fmt.Errorf("CONSPIRACY_PROBABILITY: %w", err)
	}

	return nil
}

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

	if err := loadEnvConfig(); err != nil {
		log.Fatalf("[x] Error loading env config: %v", err)
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
		if rand.Float64() < ConspiracyProbability {
			discordMessage := sendMessage(conspiracies[conspiracyIndex%len(conspiracies)], s, channelID)
			conspiracyIndex++

			go deleteMessageAfterDelay(discordMessage, s, channelID)
		}

		randomDuration = time.Duration(rand.Intn(SendMessageIntervalMax-SendMessageIntervalMin)+(SendMessageIntervalMin)) * SendMessageUnit
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
