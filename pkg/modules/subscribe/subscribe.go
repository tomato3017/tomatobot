package subscribe

import (
	"context"
	"fmt"
	"github.com/tomato3017/tomatobot/pkg/modules"
)

type SubscribeModule struct {
}

func (s *SubscribeModule) Initialize(ctx context.Context, params modules.InitializeParameters) error {
	err := params.Tomatobot.RegisterCommand("sub",
		&SubCreateCmd{tgbot: params.TgBot,
			publisher: params.Notifications,
			logger:    params.Logger})
	if err != nil {
		return fmt.Errorf("failed to register command: %w", err)
	}

	err = params.Tomatobot.RegisterCommand("sublist", &SubListCmd{publisher: params.Notifications, tgbot: params.TgBot})
	if err != nil {
		return fmt.Errorf("failed to register command: %w", err)
	}

	err = params.Tomatobot.RegisterCommand("unsub", &UnSubCmd{publisher: params.Notifications, tgbot: params.TgBot})
	if err != nil {
		return fmt.Errorf("failed to register command: %w", err)
	}

	return nil
}

func (s *SubscribeModule) Start(ctx context.Context) error {
	return nil
}

func (s *SubscribeModule) Shutdown(ctx context.Context) error {
	return nil
}
