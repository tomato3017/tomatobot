package myid

import (
	"context"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tomato3017/tomatobot/pkg/util"
)

type MyIdCmd struct {
	tgbot *tgbotapi.BotAPI
}

func (m *MyIdCmd) Execute(ctx context.Context, msg *tgbotapi.Message) error {
	_, err := m.tgbot.Send(util.NewMessagePrivate(*msg,
		fmt.Sprintf("Your ID is %d and the Chat ID is %d", msg.From.ID, msg.Chat.ID)))
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
