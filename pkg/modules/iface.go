package modules

import (
	"context"
)

type BotModule interface {
	Initialize(ctx context.Context, params InitializeParameters) error
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}
