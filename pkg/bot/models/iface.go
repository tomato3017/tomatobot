package models

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tomato3017/tomatobot/pkg/command"
)

type TomatobotInstance interface {
	RegisterCommand(name string, command command.TomatobotCommand) error
	RegisterSimpleCommand(name, desc, help string, callback func(ctx context.Context, msg *tgbotapi.Message) error) error
	RegisterChatCallback(name string, handler func(ctx context.Context, msg tgbotapi.Message)) error
}

// TODO add wrapper to tgbot to allow for easier testing and to allow interception of calls
//type TGBotCapable interface {
//	tgbotapi.BotAPI
//}
