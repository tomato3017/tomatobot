package bot

import (
	"context"
	"errors"
	"fmt"
	"github.com/tomato3017/tomatobot/pkg/bot/models"
	"github.com/tomato3017/tomatobot/pkg/modules/myid"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	"github.com/tomato3017/tomatobot/pkg/config"
	"github.com/tomato3017/tomatobot/pkg/modules"
	"github.com/tomato3017/tomatobot/pkg/modules/helloworld"
	modulemodels "github.com/tomato3017/tomatobot/pkg/modules/models"
)

type Tomatobot struct {
	cfg    config.Config
	logger zerolog.Logger
	tgbot  *tgbotapi.BotAPI

	moduleRegistry  map[string]modules.BotModule
	commandRegistry map[string]models.TomatobotCommand
	chatCallbacks   map[string]func(ctx context.Context, msg tgbotapi.Message)
}

func (t *Tomatobot) RegisterChatCallback(name string, handler func(ctx context.Context, msg tgbotapi.Message)) error {
	if _, ok := t.chatCallbacks[name]; ok {
		return fmt.Errorf("chat callback %s already registered", name)
	}

	t.chatCallbacks[name] = handler

	t.logger.Debug().Msgf("Registered chat callback: %s", name)
	return nil
}

func (t *Tomatobot) RegisterCommand(name string, commandHandler models.TomatobotCommand) error {
	t.logger.Debug().Msgf("Registering command: %s", name)
	if _, ok := t.commandRegistry[name]; ok {
		return fmt.Errorf("command %s already registered", name)
	}

	t.commandRegistry[strings.ToLower(name)] = commandHandler
	return nil
}

var _ models.TomatobotInstance = &Tomatobot{}

func (t *Tomatobot) Run(ctx context.Context) error {
	tgbot, err := tgbotapi.NewBotAPI(t.cfg.TomatoBot.Token)
	if err != nil {
		return fmt.Errorf("failed to create telegram bot: %w", err)
	}
	t.tgbot = tgbot

	tgbot.Debug = t.cfg.Debug
	t.logger.Info().Msg("Telegram bot authorized successfully")

	// Initialize modules
	for name, mod := range t.moduleRegistry {
		t.logger.Info().Msgf("Initializing module: %s", name)
		err := mod.Initialize(ctx, modulemodels.InitializeParameters{
			Cfg:       t.cfg,
			TgBot:     tgbot,
			Tomatobot: t,
			Logger:    t.logger.With().Str("module", name).Logger(),
		})
		if err != nil {
			return fmt.Errorf("failed to initialize module %s: %w", name, err)
		}
	}

	// Run main loop
	err = t.runMainLoop(ctx)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			t.logger.Info().Msg("Main loop cancelled")
		default:
			t.logger.Error().Err(err).Msg("Main loop exited with error")
			if err := t.Shutdown(ctx); err != nil {
				t.logger.Error().Err(err).Msg("Failed to shutdown bot")
			}

			return err
		}
	}

	if err := t.Shutdown(ctx); err != nil {
		t.logger.Error().Err(err).Msg("Failed to shutdown bot")
	}

	return nil
}

func (t *Tomatobot) runMainLoop(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	defer t.tgbot.StopReceivingUpdates()

	updates := t.tgbot.GetUpdatesChan(u)

	t.logger.Info().Msg("Polling started")
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled")
		case update := <-updates:
			if err := t.handleUpdate(ctx, update); err != nil {
				return fmt.Errorf("failed to handle update: %w", err)
			}
		}
	}
}

func (t *Tomatobot) handleUpdate(ctx context.Context, update tgbotapi.Update) error {
	t.logger.Trace().Msgf("Received update: %+v", update)
	if update.Message == nil {
		return nil
	}

	msg := update.Message
	if msg.IsCommand() {
		t.handleCommand(ctx, msg)
	} else {
		err := t.handleChatMessage(ctx, msg)
		if err != nil {
			return fmt.Errorf("failed to handle chat message: %w", err)
		}
	}

	return nil
}

func (t *Tomatobot) handleCommand(ctx context.Context, msg *tgbotapi.Message) {
	go func() {
		ctx, cancel := context.WithTimeout(ctx, t.cfg.TomatoBot.CommandTimeout)
		defer cancel()

		if err := t.handleCommandThread(ctx, msg); err != nil {
			// TODO error back to user
			t.logger.Error().Err(err).Msg("Failed to handle command")
		}
	}()
}

func (t *Tomatobot) handleCommandThread(ctx context.Context, msg *tgbotapi.Message) error {
	if strings.ToLower(msg.Command()) == "help" {
		return t.handleHelpCommand(ctx, msg)
	}

	command := strings.ToLower(msg.Command())
	cmdHandler, ok := t.commandRegistry[command]
	if !ok {
		return fmt.Errorf("command %s not found", command)
	}

	return cmdHandler.Execute(ctx, msg)
}

func (t *Tomatobot) handleChatMessage(ctx context.Context, msg *tgbotapi.Message) error {
	for name, handler := range t.chatCallbacks {
		t.logger.Trace().Msgf("Running chat callback: %s", name)
		handler(ctx, *msg)
	}
	return nil
}

func (t *Tomatobot) Shutdown(ctx context.Context) error {
	for name, mod := range t.moduleRegistry {
		t.logger.Debug().Msgf("Shutting down module: %s", name)
		if err := mod.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown module %s: %w", name, err)
		}
	}

	return nil
}

func (t *Tomatobot) handleHelpCommand(ctx context.Context, msg *tgbotapi.Message) error {
	helpMsg := "Available commands:\n"
	for name, cmd := range t.commandRegistry {
		helpMsg += fmt.Sprintf("/%s - %s\n", name, cmd.Description())
	}

	_, err := t.tgbot.Send(tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:           msg.Chat.ID,
			ReplyToMessageID: msg.MessageID,
		},
		Text:                  helpMsg,
		DisableWebPagePreview: false,
	})
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func NewTomatobot(cfg config.Config, logger zerolog.Logger) *Tomatobot {
	botRegistry := getModuleRegistry()

	return &Tomatobot{
		cfg:             cfg,
		logger:          logger,
		moduleRegistry:  botRegistry,
		commandRegistry: make(map[string]models.TomatobotCommand),
		chatCallbacks:   make(map[string]func(ctx context.Context, msg tgbotapi.Message)),
	}
}

func getModuleRegistry() map[string]modules.BotModule {
	return map[string]modules.BotModule{
		"helloworld": &helloworld.HelloWorldMod{},
		"myid":       &myid.MyIdMod{},
	}
}
