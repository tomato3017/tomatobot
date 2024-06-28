package myid

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	modulemodels "github.com/tomato3017/tomatobot/pkg/modules/models"
)

type MyIdMod struct {
	tgbot *tgbotapi.BotAPI
}

func (m *MyIdMod) Initialize(ctx context.Context, params modulemodels.InitializeParameters) error {
	m.tgbot = params.TgBot

	err := params.Tomatobot.RegisterCommand("myid", &MyIdCmd{tgbot: m.tgbot})
	if err != nil {
		return fmt.Errorf("failed to register command: %w", err)
	}

	return nil
}

func (m *MyIdMod) Shutdown(ctx context.Context) error {
	return nil
}
