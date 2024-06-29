package myid

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tomato3017/tomatobot/pkg/modules"
	modulemodels "github.com/tomato3017/tomatobot/pkg/modules/models"
)

type MyIdMod struct {
	tgbot *tgbotapi.BotAPI
}

var _ modules.BotModule = &MyIdMod{}

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

func (m *MyIdMod) Start(ctx context.Context) error {
	// Implementation of the Start method goes here.
	// Return nil if no error occurred during the operation.
	return nil
}
