package models

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	"github.com/tomato3017/tomatobot/pkg/bot/models"
	"github.com/tomato3017/tomatobot/pkg/config"
	"github.com/tomato3017/tomatobot/pkg/notifications"
)

type InitializeParameters struct {
	// Config is the configuration for the bot
	Cfg           config.Config
	TgBot         *tgbotapi.BotAPI
	Tomatobot     models.TomatobotInstance
	Notifications notifications.Publisher
	Logger        zerolog.Logger
}
