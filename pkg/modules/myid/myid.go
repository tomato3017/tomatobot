package myid

import (
	"context"
	"fmt"
	"github.com/tomato3017/tomatobot/pkg/bot/proxy"
	"github.com/tomato3017/tomatobot/pkg/command/models"
	"github.com/tomato3017/tomatobot/pkg/modules"
	"github.com/tomato3017/tomatobot/pkg/util"
)

type MyIdMod struct {
	botProxy proxy.TGBotImplementation
}

var _ modules.BotModule = &MyIdMod{}

func (m *MyIdMod) Initialize(ctx context.Context, params modules.InitializeParameters) error {
	m.botProxy = params.BotProxy

	err := params.Tomatobot.RegisterSimpleCommand("myid", "Gives you your user ID", "Executes the myid command",
		m.giveMyId)
	if err != nil {
		return fmt.Errorf("failed to register command: %w", err)
	}

	return nil
}

func (m *MyIdMod) giveMyId(ctx context.Context, params models.CommandParams) error {
	msg := params.Message
	_, err := m.botProxy.InnerBotAPI().Send(util.NewMessagePrivate(msg.InnerMsg(), "", fmt.Sprintf("Your ID is %d and the Chat ID is %d", msg.AssumedUserID(), msg.AssumedChatID())))
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
