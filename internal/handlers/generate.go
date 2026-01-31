package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yourusername/image-generator-bot/internal/database"
	"github.com/yourusername/image-generator-bot/internal/models"
	"github.com/yourusername/image-generator-bot/internal/services"
	"github.com/yourusername/image-generator-bot/internal/utils"
	"go.mongodb.org/mongo-driver/mongo"
)

// HandleGenerate processes /generate command
func HandleGenerate(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *mongo.Client, apiURL string) {
	userID := message.From.ID
	prompt := strings.TrimSpace(message.CommandArguments())

	// Sanitize prompt
	if len(prompt) > 500 {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Prompt too long (max 500 chars)."))
		return
	}

	// Check rate limit
	canGenerate, err := database.CheckRateLimit(ctx, db, userID)
	if err != nil {
		utils.LogError("Rate limit check failed", err)
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Internal error. Try again later."))
		return
	}
	if !canGenerate {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Rate limit exceeded: 5 generations per hour."))
		return
	}

	// Send waiting message
	waitMsg := tgbotapi.NewMessage(message.Chat.ID, "Generating your image, please wait...")
	sentMsg, err := bot.Send(waitMsg)
	if err != nil {
		utils.LogError("Failed to send waiting message", err)
		return
	}

	var imageBytes []byte
	var caption string
	var isAI bool

	if prompt == "" {
		// Fallback to Picsum
		imageBytes, err = services.FetchPicsumImage()
		caption = "Random image from Picsum"
	} else {
		// AI generation
		imageBytes, err = services.GenerateAIImage(ctx, apiURL, prompt)
		caption = fmt.Sprintf("Generated with prompt: %s", prompt)
		isAI = true
	}

	if err != nil {
		utils.LogError("Image generation failed", err)
		editMsg := tgbotapi.NewEditMessageText(message.Chat.ID, sentMsg.MessageID, "Failed to generate image. Try again.")
		bot.Send(editMsg)
		return
	}

	// Send photo
	photo := tgbotapi.NewPhoto(message.Chat.ID, tgbotapi.FileBytes{Name: "image.png", Bytes: imageBytes})
	photo.Caption = caption
	_, err = bot.Send(photo)
	if err != nil {
		utils.LogError("Failed to send photo", err)
		editMsg := tgbotapi.NewEditMessageText(message.Chat.ID, sentMsg.MessageID, "Failed to send image.")
		bot.Send(editMsg)
		return
	}

	// Delete waiting message
	bot.DeleteMessage(tgbotapi.DeleteMessageConfig{ChatID: message.Chat.ID, MessageID: sentMsg.MessageID})

	if isAI {
		// Update DB
		err = database.UpdateUserAfterGeneration(ctx, db, &models.User{UserID: userID}, prompt, time.Now())
		if err != nil {
			utils.LogError("Failed to update DB after generation", err)
		}
	}
}