package helloworld

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	"github.com/tomato3017/tomatobot/pkg/command/models"
	"github.com/tomato3017/tomatobot/pkg/modules"
	"github.com/tomato3017/tomatobot/pkg/notifications"
	"github.com/tomato3017/tomatobot/pkg/util"
	"time"
)

type HelloWorldMod struct {
	tgbot  *tgbotapi.BotAPI
	logger zerolog.Logger

	publisher notifications.Publisher
}

var _ modules.BotModule = &HelloWorldMod{}

func (h *HelloWorldMod) Initialize(ctx context.Context, params modules.InitializeParameters) error {
	h.logger = params.Logger
	h.tgbot = params.TgBot

	h.logger.Debug().Msgf("Initializing HelloWorldMod")
	err := params.Tomatobot.RegisterCommand("hello", &HelloWorldCmd{tgbot: h.tgbot})
	if err != nil {
		return fmt.Errorf("failed to register command: %w", err)
	}

	err = params.Tomatobot.RegisterSimpleCommand("hellotest", "Says hello to the world", "Executes the hello world command",
		func(ctx context.Context, params models.CommandParams) error {
			_, err := h.tgbot.Send(util.NewMessageReply(params.Message, "", "Hello, World2222!"))
			if err != nil {
				return fmt.Errorf("failed to send message: %w", err)
			}
			return nil
		})
	if err != nil {
		return fmt.Errorf("failed to register simple command: %w", err)
	}

	err = params.Tomatobot.RegisterChatCallback("helloworld_listener", h.handleChatCallback)
	if err != nil {
		return fmt.Errorf("failed to register chat callback: %w", err)
	}

	h.publisher = params.Notifications

	return nil
}

func (h *HelloWorldMod) Start(ctx context.Context) error {
	go func() {
		ticker := time.NewTicker(10 * time.Second)

		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				h.logger.Trace().Msg("Publishing hello message")
				h.publisher.Publish(notifications.Message{Topic: "helloworld.sometopic", Msg: "Hello, World!", DupeTTL: time.Minute})
			}
		}

	}()

	return nil
}

func (h *HelloWorldMod) handleChatCallback(ctx context.Context, msg tgbotapi.Message) {
	h.logger.Debug().Msgf("Got message: %s", msg.Text)
}

func (h *HelloWorldMod) Shutdown(ctx context.Context) error {
	h.logger.Debug().Msgf("Shutting down HelloWorldMod")
	return nil
}
