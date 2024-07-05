package weather

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	dbmodels "github.com/tomato3017/tomatobot/pkg/bot/models/db"
	"github.com/tomato3017/tomatobot/pkg/config"
	"github.com/tomato3017/tomatobot/pkg/modules"
	"github.com/tomato3017/tomatobot/pkg/notifications"
	"github.com/uptrace/bun"
)

type WeatherModule struct {
	cfg config.WeatherConfig

	tgbot  *tgbotapi.BotAPI
	dbConn bun.IDB

	pollingLocations []dbmodels.WeatherPollingLocation //TODO
	publisher        notifications.Publisher

	weatherPoll *poller
	logger      zerolog.Logger
}

func (w *WeatherModule) Initialize(ctx context.Context, params modules.InitializeParameters) error {
	w.logger = params.Logger
	w.cfg = params.Cfg.Modules.Weather
	w.dbConn = params.DbConn
	w.publisher = params.Notifications

	//Validate api key
	if w.cfg.APIKey == "" {
		return fmt.Errorf("no api key provided")
	}

	//Load weather polling locations
	weatherPollingLocations, err := w.getWeatherPollingLocations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get weather polling locations: %w", err)
	}
	w.pollingLocations = weatherPollingLocations

	w.logger.Trace().Any("weather_polling_locations", w.pollingLocations).Msg("Loaded weather polling locations")

	//Check subscriptions vs polling locations
	// We don't want to poll for locations that no one is subscribed to
	//TODO

	return nil
}

func (w *WeatherModule) startPolling(ctx context.Context) {
	wPoller := newPoller(pollerNewArgs{
		publisher: w.publisher,
		locations: w.pollingLocations,
		cfg:       w.cfg,
		logger:    w.logger.With().Str("thread", "weather_poller").Logger(),
	})

	wPoller.Start(ctx)

	w.weatherPoll = wPoller
}

func (w *WeatherModule) getWeatherPollingLocations(ctx context.Context) ([]dbmodels.WeatherPollingLocation, error) {
	var locations []dbmodels.WeatherPollingLocation
	err := w.dbConn.NewSelect().Model(&locations).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get weather polling locations: %w", err)
	}

	return locations, nil
}

func (w *WeatherModule) Start(ctx context.Context) error {
	w.startPolling(ctx)

	return nil
}

func (w *WeatherModule) Shutdown(ctx context.Context) error {
	w.weatherPoll.Stop()

	return nil
}
