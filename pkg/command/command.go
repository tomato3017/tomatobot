package command

import (
	"context"
	"github.com/tomato3017/tomatobot/pkg/command/models"
)

// TODO convert to parsing command text and providing to the command
// TODO command filtering based on permissions
type TomatobotCommand interface {
	BaseICommand
	Execute(ctx context.Context, params models.CommandParams) error
	Description() string
	Help() string
}

// TODO nested commands with middleware appliciation at each level
type ICommandParams interface {
	Name() string
	Args() []string
}
