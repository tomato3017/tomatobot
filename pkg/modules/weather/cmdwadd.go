package weather

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	dbmodels "github.com/tomato3017/tomatobot/pkg/bot/models/db"
	"github.com/tomato3017/tomatobot/pkg/command"
	"github.com/tomato3017/tomatobot/pkg/command/middleware"
	"github.com/tomato3017/tomatobot/pkg/command/models"
	"github.com/tomato3017/tomatobot/pkg/modules"
	"github.com/tomato3017/tomatobot/pkg/modules/weather/owm"
	"github.com/tomato3017/tomatobot/pkg/notifications"
	"github.com/tomato3017/tomatobot/pkg/util"
	"github.com/uptrace/bun"
	"strings"
)

type weatherCmdAdd struct {
	command.BaseCommand
	owmClient owm.OpenWeatherMapIClient

	dbConn    bun.IDB
	publisher notifications.Publisher
}

func newWeatherCmdAdd(params modules.InitializeParameters) *weatherCmdAdd {
	client, err := owm.NewOpenWeatherMapClient(params.Cfg.Modules.Weather.APIKey)
	if err != nil {
		panic(fmt.Errorf("failed to create OWM client: %w", err))
	}

	return &weatherCmdAdd{
		publisher:   params.Notifications,
		dbConn:      params.DbConn,
		BaseCommand: command.NewBaseCommand(middleware.WithNArgs(1)),
		owmClient:   client,
	}
}

func (w *weatherCmdAdd) Execute(ctx context.Context, params models.CommandParams) error {
	rawZipCode := params.Args[0]
	if !usZipCodeRegex.MatchString(rawZipCode) {
		return fmt.Errorf("invalid zip code format, must be 5 digits")
	}

	_, err := w.addWeatherLocation(ctx, rawZipCode, params.Message.AssumedChatID())
	if err != nil {
		return fmt.Errorf("failed to add weather location: %w", err)
	}

	_, err = params.BotProxy.Send(util.NewMessageReply(params.Message.InnerMsg(), "", "Location added successfully"))
	if err != nil {
		return fmt.Errorf("failed to send reply: %w", err)
	}

	return nil
}

func (w *weatherCmdAdd) addWeatherLocation(ctx context.Context, zipCode string, chatId int64) ([]string, error) {
	//get the geo location of the zip code

	//check if the zip code is in the db
	err := w.dbConn.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return w.insertWeatherPollerChat(ctx, tx, zipCode, chatId)
	})
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			return nil, fmt.Errorf("location already exists")
		}
		return nil, fmt.Errorf("failed to insert weather poller chat: %w", err)
	}

	topics, err := w.addLocationToSubscriptions(chatId, zipCode)
	if err != nil {
		return nil, fmt.Errorf("failed to add location to subscriptions: %w", err)
	}
	return topics, nil
}

func (w *weatherCmdAdd) populateZipGeoLoc(ctx context.Context, tx bun.IDB, zipCode string) (dbmodels.WeatherPollingLocations, error) {
	//we need to pull from the OWM API the zipcode and insert it into the database
	geoLoc, err := w.owmClient.GetLocationDataForZipCode(ctx, zipCode)
	if err != nil {
		return dbmodels.WeatherPollingLocations{}, fmt.Errorf("failed to get location data for zip code: %w", err)
	}

	weatherPollingModel := dbmodels.WeatherPollingLocations{
		Name:    geoLoc.Name,
		Country: geoLoc.Country,
		ZipCode: zipCode,
		Lon:     geoLoc.Lon,
		Lat:     geoLoc.Lat,
		Polling: true,
	}
	if _, err := tx.NewInsert().Model(&weatherPollingModel).Exec(ctx); err != nil {
		return dbmodels.WeatherPollingLocations{}, fmt.Errorf("failed to insert location into db: %w", err)
	}

	return weatherPollingModel, nil
}

func (w *weatherCmdAdd) checkZipInDb(ctx context.Context, tx bun.IDB, zipCode string) (dbmodels.WeatherPollingLocations, error) {
	geoLoc := dbmodels.WeatherPollingLocations{
		ZipCode: zipCode,
	}

	if err := tx.NewSelect().Model(&geoLoc).Where("zip_code = ?", zipCode).Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dbmodels.WeatherPollingLocations{}, nil
		}
		return dbmodels.WeatherPollingLocations{}, fmt.Errorf("failed to check if zip code is in db: %w", err)
	}

	return geoLoc, nil
}

func (w *weatherCmdAdd) Description() string {
	return "Add a location to the weather alerting"
}

func (w *weatherCmdAdd) Help() string {
	return "/weather add <zipcode> - Add a location to the weather alerting"
}

func (w *weatherCmdAdd) addLocationToSubscriptions(chatId int64, zipCode string) ([]string, error) {
	topics := make([]string, 0)

	for _, topic := range weatherPublisherEventTypes {
		topicName := topic.fullTopicPath(zipCode)

		_, err := w.publisher.Subscribe(notifications.Subscriber{
			ChatId:       chatId,
			TopicPattern: topicName,
		})
		if err != nil {
			if errors.Is(err, notifications.ErrSubExists) {
				topics = append(topics, topicName)
				continue
			}

			return nil, fmt.Errorf("failed to subscribe to topic: %w", err)
		}
		topics = append(topics, topicName)
	}

	return topics, nil
}

func (w *weatherCmdAdd) insertWeatherPollerChat(ctx context.Context, tx bun.Tx, zipCode string, chatId int64) error {
	weatherPollingModel, err := w.checkZipInDb(ctx, tx, zipCode)
	if err != nil {
		return fmt.Errorf("failed to check zip in db: %w", err)
	}

	if weatherPollingModel.IsEmpty() {
		//Insert a row in the db AND get the geo location
		newWeatherPollMdl, err := w.populateZipGeoLoc(ctx, tx, zipCode)
		if err != nil {
			return fmt.Errorf("failed to populate zip geo location: %w", err)
		}

		weatherPollingModel = newWeatherPollMdl
	} else if !weatherPollingModel.Polling {
		//update the polling to true
		if _, err := tx.NewUpdate().Model(&weatherPollingModel).Set("polling = ?", true).WherePK().Exec(ctx); err != nil {
			return fmt.Errorf("failed to update polling: %w", err)
		}
	}

	//Insert a new entry into the weather polling chats
	weatherPollerChat := dbmodels.WeatherPollerChats{
		ChatID:           chatId,
		PollerLocationID: weatherPollingModel.ID,
	}

	if _, err := tx.NewInsert().Model(&weatherPollerChat).Exec(ctx); err != nil {
		return fmt.Errorf("failed to insert weather poller chat: %w", err)
	}

	return nil
}
