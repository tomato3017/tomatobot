package myid

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tomato3017/tomatobot/pkg/command"
	"github.com/tomato3017/tomatobot/pkg/modules"
	"github.com/tomato3017/tomatobot/pkg/util"
)

type MyIdMod struct {
	tgbot *tgbotapi.BotAPI
}

var _ modules.BotModule = &MyIdMod{}

func (m *MyIdMod) Initialize(ctx context.Context, params modules.InitializeParameters) error {
	m.tgbot = params.TgBot

	err := params.Tomatobot.RegisterSimpleCommand("myid", "Gives you your user ID", "Executes the myid command",
		m.giveMyId)
	if err != nil {
		return fmt.Errorf("failed to register command: %w", err)
	}

	return nil
}

func (m *MyIdMod) giveMyId(ctx context.Context, params command.CommandParams) error {
	msg := params.Message
	_, err := m.tgbot.Send(util.NewMessagePrivate(*msg,
		fmt.Sprintf("Your ID is %d and the Chat ID is %d", msg.From.ID, msg.Chat.ID)))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
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
