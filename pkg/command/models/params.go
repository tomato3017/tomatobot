package models

import (
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tomato3017/tomatobot/pkg/bot/proxy"
)

type CommandParams struct {
	CommandName string
	Args        []string
	Message     *tgbotapi.Message
	BotProxy    proxy.TGBotImplementation
}
