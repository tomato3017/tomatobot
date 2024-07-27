package weather

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	dbmodels "github.com/tomato3017/tomatobot/pkg/bot/models/db"
	"github.com/tomato3017/tomatobot/pkg/command"
	"github.com/tomato3017/tomatobot/pkg/command/middleware"
	"github.com/tomato3017/tomatobot/pkg/command/models"
	"github.com/tomato3017/tomatobot/pkg/modules"
	"github.com/tomato3017/tomatobot/pkg/notifications"
	"github.com/tomato3017/tomatobot/pkg/util"
	"github.com/uptrace/bun"
)

type weatherCmdRemove struct {
	command.BaseCommand
	dbConn    bun.IDB
	logger    zerolog.Logger
	publisher notifications.Publisher
}

func newWeatherCmdRemove(params modules.InitializeParameters) *weatherCmdRemove {
	return &weatherCmdRemove{
		dbConn:      params.DbConn,
		BaseCommand: command.NewBaseCommand(middleware.WithNArgs(1)),
		logger:      params.Logger,
		publisher:   params.Notifications,
	}
}

func (w *weatherCmdRemove) Execute(ctx context.Context, params models.CommandParams) error {
	zipCode := params.Args[0]
	if !usZipCodeRegex.MatchString(zipCode) {
		return fmt.Errorf("invalid zip code format, must be 5 digits")
	}

	w.logger.Debug().Str("zip_code", zipCode).Int64("chat_id", params.Message.Chat.ID).Msg("Removing location")
	err := w.removeLocationInDb(ctx, zipCode, params.Message.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to remove location: %w", err)
	}

	err = w.removeLocationFromSubscriptions(params.Message.Chat.ID, zipCode)
	if err != nil {
		return fmt.Errorf("failed to remove subscriptions: %w", err)
	}

	_, err = params.BotProxy.Send(util.NewMessageReply(params.Message, "", "Location removed successfully"))
	if err != nil {
		return fmt.Errorf("failed to send reply: %w", err)
	}

	return nil
}

func (w *weatherCmdRemove) removeLocationInDb(ctx context.Context, zipCode string, chatID int64) error {
	var chats []dbmodels.WeatherPollerChats
	err := w.dbConn.NewSelect().
		Model(&chats).
		Where("chat_id = ?", chatID).
		Join("JOIN weather_polling_locations wpl").
		JoinOn("wpl.id = weather_poller_chats.poller_location_id").
		Where("wpl.zip_code = ?", zipCode).Scan(ctx)

	if err != nil {
		return fmt.Errorf("failed to get chat: %w", err)
	}

	err = w.dbConn.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for _, chat := range chats {
			w.logger.Debug().Int64("chat_id", chat.ChatID).Int("location_id", chat.PollerLocationID).Msg("Removing chat")
			_, err := tx.NewDelete().Model(&chat).WherePK().Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to delete chat: %w", err)
			}
		}

		count, err := tx.NewSelect().Model(&dbmodels.WeatherPollerChats{}).
			Join("JOIN weather_polling_locations wpl").
			JoinOn("wpl.id = weather_poller_chats.poller_location_id").
			Where("wpl.zip_code = ?", zipCode).Count(ctx)
		if err != nil {
			return fmt.Errorf("failed to count locations: %w", err)
		}
		if count == 0 {
			_, err = tx.NewUpdate().
				Model(&dbmodels.WeatherPollingLocations{}).
				Set("polling = ?", false).
				Where("zip_code = ?", zipCode).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to update location: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to run transaction: %w", err)
	}

	return nil
}

func (w *weatherCmdRemove) removeLocationFromSubscriptions(chatID int64, zipCode string) error {
	//get all the subscriptions for the chat
	subs, err := w.publisher.GetSubscriptions(chatID)
	if err != nil {
		return fmt.Errorf("failed to get subscriptions: %w", err)
	}

	for _, sub := range subs {
		for _, eventType := range weatherPublisherEventTypes {
			topic := eventType.fullTopicPath(zipCode)
			if sub.TopicPattern == topic {
				err := w.publisher.Unsubscribe(sub.ID, chatID)
				if err != nil {
					return fmt.Errorf("failed to unsubscribe: %w", err)
				}
			}
		}
	}

	return nil
}

func (w *weatherCmdRemove) Description() string {
	return "Remove a weather location"
}

func (w *weatherCmdRemove) Help() string {
	return "/weather remove <zip>"
}
