package handlers

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yourusername/image-generator-bot/internal/database"
	"github.com/yourusername/image-generator-bot/internal/utils"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandleHistory(bot *tgbotapi.BotAPI, message *tgbotapi.Message, db *mongo.Client) {
	ctx := context.Background()
	histories, err := database.GetUserHistory(ctx, db, message.From.ID, 10)
	if err != nil {
		utils.LogError("Failed to get history", err)
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Failed to retrieve history."))
		return
	}

	if len(histories) == 0 {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "No generation history."))
		return
	}

	var sb strings.Builder
	sb.WriteString("Your last generations:\n")
	for i, h := range histories {
		sb.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, h.Prompt, h.Timestamp.Format(time.RFC822)))
	}

	bot.Send(tgbotapi.NewMessage(message.Chat.ID, sb.String()))
}