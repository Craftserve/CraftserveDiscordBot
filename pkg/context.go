package pkg

import (
	"context"
	"csrvbot/pkg/logger"
)

func CreateContext() context.Context {
	ctx := context.Background()
	ctx = logger.ContextWithLogger(ctx, logger.GetLoggerFromContext(ctx))
	return ctx
}
