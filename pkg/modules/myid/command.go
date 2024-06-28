package myid

import (
	"context"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MyIdCmd struct {
	tgbot *tgbotapi.BotAPI
}

func (m *MyIdCmd) Execute(ctx context.Context, msg *tgbotapi.Message) error {
	_, err := m.tgbot.Send(tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:           msg.Chat.ID,
			ReplyToMessageID: msg.MessageID,
		},
		Text:                  fmt.Sprintf("Your ID is %d and the Chat ID is %d", msg.From.ID, msg.Chat.ID),
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
