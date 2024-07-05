package bot

import (
	"context"
	"errors"
	"fmt"
	"github.com/tomato3017/tomatobot/pkg/bot/models"
	"github.com/tomato3017/tomatobot/pkg/command"
	models2 "github.com/tomato3017/tomatobot/pkg/command/models"
	"github.com/tomato3017/tomatobot/pkg/db"
	"github.com/tomato3017/tomatobot/pkg/modules/myid"
	"github.com/tomato3017/tomatobot/pkg/modules/topic"
	"github.com/tomato3017/tomatobot/pkg/modules/weather"
	"github.com/tomato3017/tomatobot/pkg/notifications"
	"github.com/tomato3017/tomatobot/pkg/sqlmigrate"
	"github.com/tomato3017/tomatobot/pkg/util"
	"github.com/uptrace/bun"
	"slices"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	"github.com/tomato3017/tomatobot/pkg/config"
	"github.com/tomato3017/tomatobot/pkg/modules"
	"github.com/tomato3017/tomatobot/pkg/modules/helloworld"
)

type Tomatobot struct {
	cfg    config.Config
	logger zerolog.Logger
	tgbot  *tgbotapi.BotAPI

	moduleRegistry  map[string]modules.BotModule
	loadedModules   map[string]modules.BotModule
	commandRegistry map[string]command.TomatobotCommand
	chatCallbacks   map[string]func(ctx context.Context, msg tgbotapi.Message)

	notiPublisher *notifications.NotificationPublisher

	dbConn *bun.DB
}

var _ models.TomatobotInstance = &Tomatobot{}

func (t *Tomatobot) RegisterChatCallback(name string, handler func(ctx context.Context, msg tgbotapi.Message)) error {
	if _, ok := t.chatCallbacks[name]; ok {
		return fmt.Errorf("chat callback %s already registered", name)
	}

	t.chatCallbacks[name] = handler

	t.logger.Debug().Msgf("Registered chat callback: %s", name)
	return nil
}

func (t *Tomatobot) RegisterCommand(name string, commandHandler command.TomatobotCommand) error {
	t.logger.Debug().Msgf("Registering command: %s", name)
	if _, ok := t.commandRegistry[name]; ok {
		return fmt.Errorf("command %s already registered", name)
	}

	t.commandRegistry[strings.ToLower(name)] = commandHandler
	return nil
}

var _ models.TomatobotInstance = &Tomatobot{}

func (t *Tomatobot) Run(ctx context.Context) error {
	// Get the DB connection
	err := t.openDbConnection(ctx)
	if err != nil {
		return err
	}
	defer util.CloseSafely(t.dbConn)

	t.logger.Debug().Msg("Database connection successful")

	tgbot, err := tgbotapi.NewBotAPI(t.cfg.TomatoBot.Token)
	if err != nil {
		return fmt.Errorf("failed to create telegram bot: %w", err)
	}
	t.tgbot = tgbot

	tgbot.Debug = t.cfg.Debug
	t.logger.Info().Msg("Telegram bot authorized successfully")

	// Initialize the notification publisher
	t.notiPublisher = notifications.NewNotificationPublisher(tgbot, t.dbConn,
		notifications.WithLogger(t.logger.With().Str("module", "notifications").Logger()))

	// Initialize modules
	err = t.initializeModules(ctx, tgbot)
	if err != nil {
		return err
	}

	// Start notification publisher
	err = t.notiPublisher.Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start notification publisher: %w", err)
	}
	defer util.CloseSafely(t.notiPublisher)

	for name, module := range t.loadedModules {
		t.logger.Trace().Msgf("Starting module: %s", name)
		err := module.Start(ctx)
		if err != nil {
			return fmt.Errorf("failed to start module %s: %w", name, err)
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

func (t *Tomatobot) openDbConnection(ctx context.Context) error {
	t.logger.Trace().Msg("Getting DB connection")
	dbConn, err := db.GetDbConnection(t.cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to get DB connection: %w", err)
	}
	t.dbConn = dbConn

	t.logger.Debug().Msg("Migrating DB schema")
	numMigrations, err := sqlmigrate.MigrateDbSchema(ctx, dbConn)
	if err != nil {
		return fmt.Errorf("failed to migrate DB schema: %w", err)
	}

	t.logger.Debug().Msgf("DB Migrations successful. %d migrations applied", numMigrations)

	return nil
}

func (t *Tomatobot) initializeModules(ctx context.Context, tgbot *tgbotapi.BotAPI) error {
	for name, mod := range t.moduleRegistry {
		if t.cfg.TomatoBot.AllModules != nil && !*t.cfg.TomatoBot.AllModules {
			if !slices.Contains(t.cfg.ModulesToLoad, name) {
				t.logger.Debug().Msgf("Skipping module: %s", name)
				continue
			}
		}
		t.logger.Info().Msgf("Initializing module: %s", name)
		err := mod.Initialize(ctx, modules.InitializeParameters{
			Cfg:           t.cfg,
			TgBot:         tgbot,
			Tomatobot:     t,
			Logger:        t.logger.With().Str("module", name).Logger(),
			Notifications: t.notiPublisher,
			DbConn:        t.dbConn,
		})
		if err != nil {
			return fmt.Errorf("failed to initialize module %s: %w", name, err)
		}

		t.loadedModules[name] = mod
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
			t.logger.Error().Err(err).Msg("Failed to handle command")
			t.tgbot.Send(tgbotapi.MessageConfig{
				BaseChat: tgbotapi.BaseChat{
					ChatID:           msg.Chat.ID,
					ReplyToMessageID: msg.MessageID,
				},
				Text:                  fmt.Sprintf("Error: %s", err.Error()),
				DisableWebPagePreview: false,
			})
		}
	}()
}

func (t *Tomatobot) handleCommandThread(ctx context.Context, msg *tgbotapi.Message) error {
	if strings.ToLower(msg.Command()) == "help" {
		return t.handleHelpCommand(ctx, msg)
	}

	msgCommand := strings.ToLower(msg.Command())
	cmdHandler, ok := t.commandRegistry[msgCommand]
	if !ok {
		return fmt.Errorf("command %s not found", msgCommand)
	}

	args := strings.Split(msg.CommandArguments(), " ")
	params := models2.CommandParams{
		CommandName: msgCommand,
		Args:        args,
		Message:     msg,
		TgBot:       t.tgbot,
	}

	return cmdHandler.Execute(ctx, params)
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

func (t *Tomatobot) RegisterSimpleCommand(name, desc, help string, callback command.CommandCallback) error {
	cmd := command.NewSimpleCommand(callback, desc, help)
	return t.RegisterCommand(name, cmd)
}

func NewTomatobot(cfg config.Config, logger zerolog.Logger) *Tomatobot {
	botRegistry := getModuleRegistry()

	return &Tomatobot{
		cfg:             cfg,
		logger:          logger,
		moduleRegistry:  botRegistry,
		loadedModules:   make(map[string]modules.BotModule),
		commandRegistry: make(map[string]command.TomatobotCommand),
		chatCallbacks:   make(map[string]func(ctx context.Context, msg tgbotapi.Message)),
	}
}

func getModuleRegistry() map[string]modules.BotModule {
	return map[string]modules.BotModule{
		"helloworld": &helloworld.HelloWorldMod{},
		"myid":       &myid.MyIdMod{},
		"topic":      &topic.TopicModule{},
		"weather":    &weather.WeatherModule{},
	}
}
