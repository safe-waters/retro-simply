package logger

import (
	"context"
	"os"

	"github.com/safe-waters/retro-simply/backend/pkg/user"
	"github.com/safe-waters/structured_logger"
)

const (
	idKey  = "correlationId"
	msgKey = "message"
)

var l = structured_logger.New(structured_logger.DefaultSkip+1, os.Stdout)

func Info(ctx context.Context, msg string) {
	u, _ := user.FromContext(ctx)
	l.Info(
		[2]string{idKey, u.CorrelationId},
		[2]string{msgKey, msg},
	)
}

func Error(ctx context.Context, err error) {
	u, _ := user.FromContext(ctx)
	l.Error(
		[2]string{idKey, u.CorrelationId},
		[2]string{msgKey, err.Error()},
	)
}
