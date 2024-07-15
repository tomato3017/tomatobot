package weather

import (
	"github.com/tomato3017/tomatobot/pkg/command"
	"github.com/tomato3017/tomatobot/pkg/command/middleware"
	"github.com/tomato3017/tomatobot/pkg/modules"
)

// /weather add <zip>
// /weather remove <zip>
// /weather list

type weatherCommand struct {
	command.BaseCommand
}

func newWeatherCommand(params modules.InitializeParameters) (*weatherCommand, error) {
	weatherCmd := &weatherCommand{
		BaseCommand: command.NewBaseCommand(middleware.WithAdminPermission()),
	}

	err := weatherCmd.RegisterSubcommand("add", newWeatherCmdAdd(params))
	if err != nil {
		return nil, err
	}

	err = weatherCmd.RegisterSubcommand("remove", newWeatherCmdRemove(params))
	if err != nil {
		return nil, err
	}

	err = weatherCmd.RegisterSubcommand("list", newWeatherCmdList(params))
	if err != nil {
		return nil, err
	}

	return weatherCmd, nil
}

func (w *weatherCommand) Description() string {
	return "Weather related commands"
}

func (w *weatherCommand) Help() string {
	return "/weather"
}
