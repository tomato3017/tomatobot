package models

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TomatobotInstance interface {
	RegisterCommand(name string, command TomatobotCommand) error
	RegisterChatCallback(name string, handler func(ctx context.Context, msg tgbotapi.Message)) error
}

// TODO convert to parsing command text and providing to the command
type TomatobotCommand interface {
	Execute(ctx context.Context, msg *tgbotapi.Message) error
	Description() string
	Help() string
}

// TODO add wrapper to tgbot to allow for easier testing and to allow interception of calls
//type TGBotCapable interface {
//	tgbotapi.BotAPI
//}
