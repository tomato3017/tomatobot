package models

import "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type CommandParams struct {
	CommandName string
	Args        []string
	Message     *tgbotapi.Message
	TgBot       *tgbotapi.BotAPI
}
