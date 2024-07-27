package middleware

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tomato3017/tomatobot/pkg/command/models"
)

type MiddlewareFunc func(ctx context.Context, params models.CommandParams) error

func WithNArgs(n int) MiddlewareFunc {
	return func(ctx context.Context, params models.CommandParams) error {
		if len(params.Args) != n {
			return fmt.Errorf("expected %d arguments, got %d", n, len(params.Args))
		}

		return nil
	}
}

func WithMinArgs(min int) MiddlewareFunc {
	return func(ctx context.Context, params models.CommandParams) error {
		if len(params.Args) < min {
			return fmt.Errorf("expected at least %d arguments, got %d", min, len(params.Args))
		}

		return nil
	}
}

func WithMaxArgs(max int) MiddlewareFunc {
	return func(ctx context.Context, params models.CommandParams) error {
		if len(params.Args) > max {
			return fmt.Errorf("expected at most %d arguments, got %d", max, len(params.Args))
		}

		return nil
	}
}

func WithAdminPermission() MiddlewareFunc {
	return func(ctx context.Context, params models.CommandParams) error {
		if !(params.Message.Chat.IsGroup() || params.Message.Chat.IsSuperGroup()) {
			return nil
		}

		//TODO caching
		administrators, err := params.BotProxy.InnerBotAPI().GetChatAdministrators(tgbotapi.ChatAdministratorsConfig{ //nolint:govet
			tgbotapi.ChatConfig{ChatID: params.Message.Chat.ID},
		})
		if err != nil {
			return fmt.Errorf("failed to get chat administrators: %w", err)
		}

		for _, administrator := range administrators {
			if administrator.User.ID == params.Message.From.ID {
				return nil
			}
		}

		return fmt.Errorf("you are not an administrator")
	}
}

func WithUserId(userId int64) MiddlewareFunc {
	return func(ctx context.Context, params models.CommandParams) error {
		if params.Message.From.ID != userId {
			return fmt.Errorf("you are not authorized to use this command")
		}

		return nil
	}
}
