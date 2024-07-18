package modules

import (
	"context"
)

// TODO rename
type BotModule interface {
	Initialize(ctx context.Context, params InitializeParameters) error
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}
