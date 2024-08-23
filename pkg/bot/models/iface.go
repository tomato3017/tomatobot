package models

import (
	"context"
	"github.com/tomato3017/tomatobot/pkg/bot/models/tgapi"
	"github.com/tomato3017/tomatobot/pkg/command"
)

type TomatobotInstance interface {
	RegisterCommand(name string, command command.TomatobotCommand) error
	RegisterSimpleCommand(name, desc, help string, callback command.CommandCallback) error
	RegisterChatCallback(name string, handler func(ctx context.Context, msg tgapi.TGBotMsg)) error
}
