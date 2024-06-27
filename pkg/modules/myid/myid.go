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

type MyIdCmd struct {
	tgbot *tgbotapi.BotAPI
}

func (m *MyIdCmd) Execute(ctx context.Context, msg *tgbotapi.Message) error {
	_, err := m.tgbot.Send(tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:           msg.Chat.ID,
			ReplyToMessageID: msg.MessageID,
		},
		Text:                  fmt.Sprintf("Your ID is %d", msg.From.ID),
		DisableWebPagePreview: false,
	})
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

func (m *MyIdCmd) Description() string {
	return "Returns the ID of the user who sent the command"
}

func (m *MyIdCmd) Help() string {
	return "Executes the myid command"
}
