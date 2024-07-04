package topic

import (
	"context"
	"fmt"
	"github.com/tomato3017/tomatobot/pkg/modules"
)

type TopicModule struct {
}

func (s *TopicModule) Initialize(ctx context.Context, params modules.InitializeParameters) error {
	topicCmd, err := newTopicCmd(params.Notifications, params.TgBot, params.Logger)
	if err != nil {
		return fmt.Errorf("failed to create command: %w", err)
	}

	err = params.Tomatobot.RegisterCommand("topic", topicCmd)
	if err != nil {
		return fmt.Errorf("failed to register command: %w", err)
	}

	return nil
}

func (s *TopicModule) Start(ctx context.Context) error {
	return nil
}

func (s *TopicModule) Shutdown(ctx context.Context) error {
	return nil
}
