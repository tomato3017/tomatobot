package models

import (
	"github.com/tomato3017/tomatobot/pkg/bot/models/tgapi"
	"github.com/tomato3017/tomatobot/pkg/bot/proxy"
)

type CommandParams struct {
	CommandName string
	Args        []string
	Message     tgapi.TGBotMsg
	BotProxy    proxy.TGBotImplementation
}
