package helloworld

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tomato3017/tomatobot/pkg/command"
	"github.com/tomato3017/tomatobot/pkg/command/models"
)

type HelloWorldCmd struct {
	command.BaseCommand
	tgbot *tgbotapi.BotAPI
}

var _ command.TomatobotCommand = &HelloWorldCmd{}

func (h *HelloWorldCmd) Execute(ctx context.Context, params models.CommandParams) error {
	msg := params.Message
	_, err := h.tgbot.Send(tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:           msg.Chat.ID,
			ReplyToMessageID: msg.MessageID,
		},
		Text:                  "Hello World!",
		DisableWebPagePreview: false,
	})
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

func (h *HelloWorldCmd) Description() string {
	return "Says hello to the world"
}

func (h *HelloWorldCmd) Help() string {
	return "Executes the hello world command"
}

func NewHelloWorldCmd(tgbot *tgbotapi.BotAPI) *HelloWorldCmd {
	return &HelloWorldCmd{
		BaseCommand: command.NewBaseCommand(),
		tgbot:       tgbot,
	}
}
