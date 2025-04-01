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

var (
	// CommitHash and BuildData get replaced during build with the commit hash and time of build.
	CommitHash string
	BuildDate  string
	// StartTime is reset every time the application is restarted
	StartTime = time.Now()
)

// Embed JSON files at build time

//go:embed praises.json
var praisesFile embed.FS

//go:embed conspiracies.json
var conspiraciesFile embed.FS

var (
	// messages sourced from the embedded files
	praises      []string
	conspiracies []string

	// internally track which messages have been sent
	praiseIndex     int
	conspiracyIndex int
)

// Settings for bot sourced from env variables
var (
	botToken  string
	channelID string

	SendMessageIntervalMin int
	SendMessageIntervalMax int
	SendMessageUnit        time.Duration

	DeleteConspiracyDelay time.Duration
	ConspiracyProbability float64
)

func main() {
	if err := loadEnvConfig(); err != nil {
		log.Fatalf("[x] Failed to load configuration: %v", err)
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

	dg.AddHandler(readyHandler)
	interactionHandlers := make(map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate))
	for _, interaction := range interactions {
		interactionHandlers[interaction.command.Name] = interaction.handler
	}
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := interactionHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	// Open WebSocket connection
	if err := dg.Open(); err != nil {
		log.Fatalf("[x] Error opening connection: %v", err)
	}
	defer dg.Close()

	for _, entry := range interactions {
		created, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", entry.command)
		if err != nil {
			log.Fatalf("Cannot create '%s' command: %v", entry.command.Name, err)
		}
		interactionHandlers[created.Name] = entry.handler
	}

	// Start the message scheduler
	go startScheduler(dg, channelID)

	// Graceful shutdown handling
	log.Println("[✓] Swiftspiracy Bot is now running. Press CTRL+C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("[↓] Shutting down bot gracefully...")
}

func loadEnvConfig() error {
	if err := godotenv.Load(); err != nil {
		log.Println("[!] No .env file found, using system environment variables.")
	}

	botToken = os.Getenv("DISCORD_BOT_TOKEN")
	if botToken == "" {
		return fmt.Errorf("Missing DISCORD_BOT_TOKEN")
	}
	channelID = os.Getenv("DISCORD_CHANNEL_ID")
	if botToken == "" {
		return fmt.Errorf("Missing DISCORD_CHANNEL_ID")
	}

	var err error

	SendMessageIntervalMin, err = strconv.Atoi(os.Getenv("SEND_MESSAGE_INTERVAL_MIN"))
	if err != nil {
		return fmt.Errorf("SEND_MESSAGE_INTERVAL_MIN: %w", err)
	}
	SendMessageIntervalMax, err = strconv.Atoi(os.Getenv("SEND_MESSAGE_INTERVAL_MAX"))
	if err != nil {
		return fmt.Errorf("SEND_MESSAGE_INTERVAL_MAX: %w", err)
	}

	switch os.Getenv("SEND_MESSAGE_UNIT") {
	case "second", "seconds":
		SendMessageUnit = time.Second
	case "minute", "minutes":
		SendMessageUnit = time.Minute
	case "millisecond", "milliseconds":
		SendMessageUnit = time.Millisecond
	default:
		return fmt.Errorf("unsupported SEND_MESSAGE_UNIT: %s", os.Getenv("SEND_MESSAGE_UNIT"))
	}

	DeleteConspiracyDelay, err = time.ParseDuration(os.Getenv("DELETE_CONSPIRACY_DELAY"))
	if err != nil {
		return fmt.Errorf("DELETE_CONSPIRACY_DELAY: %w", err)
	}

	ConspiracyProbability, err = strconv.ParseFloat(os.Getenv("CONSPIRACY_PROBABILITY"), 64)
	if err != nil {
		return fmt.Errorf("CONSPIRACY_PROBABILITY: %w", err)
	}

	return nil
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

var interactions = []struct {
	command *discordgo.ApplicationCommand
	handler func(s *discordgo.Session, i *discordgo.InteractionCreate)
}{
	{
		command: &discordgo.ApplicationCommand{
			Name:        "buildinfo",
			Description: "Get the bot's build info and uptime",
		},
		handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			uptime := time.Since(StartTime).Round(time.Second)
			response := fmt.Sprintf(
				":tools: Build Info:\n- Commit Hash: `%s`\n- Build Date: `%s`\n- Uptime: `%s`",
				CommitHash,
				BuildDate,
				uptime,
			)
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: response,
				},
			})
			if err != nil {
				log.Printf("[x] Failed to respond to interaction: %v", err)
			}
		},
	},
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
