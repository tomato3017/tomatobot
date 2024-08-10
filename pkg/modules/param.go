package modules

import (
	"github.com/rs/zerolog"
	"github.com/tomato3017/tomatobot/pkg/bot/models"
	"github.com/tomato3017/tomatobot/pkg/bot/proxy"
	"github.com/tomato3017/tomatobot/pkg/config"
	"github.com/tomato3017/tomatobot/pkg/notifications"
	"github.com/uptrace/bun"
)

type InitializeParameters struct {
	// Config is the configuration for the bot
	Cfg           config.Config
	BotProxy      proxy.TGBotImplementation
	Tomatobot     models.TomatobotInstance
	Notifications notifications.Publisher
	DbConn        bun.IDB
	Logger        zerolog.Logger
}
