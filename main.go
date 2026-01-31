package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/animestopbus-ctrl/image-generator-bot/internal/database"
	"github.com/animestopbus-ctrl/image-generator-bot/internal/handlers"
	"github.com/animestopbus-ctrl/image-generator-bot/internal/utils"
)

func main() {
	config := utils.LoadConfig()

	// Init MongoDB
	dbClient, err := database.NewMongoClient(config.MongoURI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer dbClient.Disconnect(context.Background())

	// Init bot
	bot, err := tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}
	bot.Debug = false

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for update := range updates {
			if update.Message == nil {
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			switch update.Message.Command() {
			case "start":
				handlers.HandleStart(bot, update.Message)
			case "generate":
				handlers.HandleGenerate(ctx, bot, update.Message, dbClient, config.PythonAPIURL)
			case "history":
				handlers.HandleHistory(bot, update.Message, dbClient)
			default:
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command. Use /start, /generate, or /history."))
			}
		}
	}()

	<-sigChan
	log.Println("Shutting down...")
}
