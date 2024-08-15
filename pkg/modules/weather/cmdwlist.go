package weather

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	dbmodels "github.com/tomato3017/tomatobot/pkg/bot/models/db"
	"github.com/tomato3017/tomatobot/pkg/command"
	"github.com/tomato3017/tomatobot/pkg/command/middleware"
	"github.com/tomato3017/tomatobot/pkg/command/models"
	"github.com/tomato3017/tomatobot/pkg/modules"
	"github.com/tomato3017/tomatobot/pkg/util"
	"github.com/uptrace/bun"
	"strings"
)

type weatherCmdList struct {
	command.BaseCommand
	dbConn bun.IDB
}

func newWeatherCmdList(params modules.InitializeParameters) *weatherCmdList {
	return &weatherCmdList{
		BaseCommand: command.NewBaseCommand(middleware.WithNArgs(0)),
		dbConn:      params.DbConn,
	}
}

func (w *weatherCmdList) Execute(ctx context.Context, params models.CommandParams) error {
	//get the list of locations
	dbLocations, err := w.getLocations(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to get locations: %w", err)
	}

	if len(dbLocations) == 0 {
		_, err = params.BotProxy.Send(util.NewMessageReply(params.Message.InnerMsg(), "", "No locations found"))
		if err != nil {
			return fmt.Errorf("failed to send reply: %w", err)
		}
		return nil
	}

	outStr := strings.Builder{}
	outStr.WriteString("Weather locations:\n")
	for _, loc := range dbLocations {
		outStr.WriteString(fmt.Sprintf("%s \\- %s, %s\n", loc.ZipCode, loc.Name, loc.Country))
	}

	_, err = params.BotProxy.Send(util.NewMessageReply(params.Message.InnerMsg(), tgbotapi.ModeMarkdownV2, outStr.String()))
	if err != nil {
		return fmt.Errorf("failed to send reply: %w", err)
	}

	return nil
}

func (w *weatherCmdList) getLocations(ctx context.Context, params models.CommandParams) ([]dbmodels.WeatherPollingLocations, error) {
	//get the list of locations
	var dbLocations []dbmodels.WeatherPollingLocations
	err := w.dbConn.NewSelect().Model(&dbLocations).
		Join("JOIN weather_poller_chats wp").
		JoinOn("wp.poller_location_id = weather_polling_locations.id").
		Where("wp.chat_id = ?", params.Message.AssumedChatID()).
		Where("weather_polling_locations.polling=?", true).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return dbLocations, nil
}

func (w *weatherCmdList) Description() string {
	return "List all weather locations"
}

func (w *weatherCmdList) Help() string {
	return "List all weather locations"
}
