package helloworld

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	"github.com/tomato3017/tomatobot/pkg/modules"
	modulemodels "github.com/tomato3017/tomatobot/pkg/modules/models"
)

type HelloWorldMod struct {
	tgbot  *tgbotapi.BotAPI
	logger zerolog.Logger
}

var _ modules.BotModule = &HelloWorldMod{}

func (h *HelloWorldMod) Initialize(ctx context.Context, params modulemodels.InitializeParameters) error {
	h.logger = params.Logger
	h.tgbot = params.TgBot

	h.logger.Debug().Msgf("Initializing HelloWorldMod")
	err := params.Tomatobot.RegisterCommand("hello", &HelloWorldCmd{tgbot: h.tgbot})
	if err != nil {
		return fmt.Errorf("failed to register command: %w", err)
	}

	err = params.Tomatobot.RegisterChatCallback("helloworld_listener", h.handleChatCallback)
	if err != nil {
		return fmt.Errorf("failed to register chat callback: %w", err)
	}

	return nil
}

func (h *HelloWorldMod) handleChatCallback(ctx context.Context, msg tgbotapi.Message) {
	h.logger.Debug().Msgf("Got message: %s", msg.Text)
}

func (h *HelloWorldMod) Shutdown(ctx context.Context) error {
	h.logger.Debug().Msgf("Shutting down HelloWorldMod")
	return nil
}
