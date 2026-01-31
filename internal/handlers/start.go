package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleStart(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Welcome to AI Image Generator Bot!\n\nCommands:\n/start - This message\n/generate [prompt] - Generate AI image or random if no prompt\n/history - Your last generations")
	bot.Send(msg)
}