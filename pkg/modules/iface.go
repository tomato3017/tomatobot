package modules

import (
	"context"
	modulemodels "github.com/tomato3017/tomatobot/pkg/modules/models"
)

type BotModule interface {
	Initialize(ctx context.Context, params modulemodels.InitializeParameters) error
	Shutdown(ctx context.Context) error
}
