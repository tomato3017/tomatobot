package models

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tomato3017/tomatobot/pkg/command"
)

type TomatobotInstance interface {
	RegisterCommand(name string, command command.TomatobotCommand) error
	RegisterSimpleCommand(name, desc, help string, callback command.CommandCallback) error
	RegisterChatCallback(name string, handler func(ctx context.Context, msg tgbotapi.Message)) error
}
