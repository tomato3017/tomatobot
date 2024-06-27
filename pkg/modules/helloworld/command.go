package helloworld

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tomato3017/tomatobot/pkg/bot/models"
)

type HelloWorldCmd struct {
	tgbot *tgbotapi.BotAPI
}

var _ models.TomatobotCommand = &HelloWorldCmd{}

func (h *HelloWorldCmd) Execute(ctx context.Context, msg *tgbotapi.Message) error {

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