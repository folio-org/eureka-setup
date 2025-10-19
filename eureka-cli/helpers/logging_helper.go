package helpers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/folio-org/eureka-cli/action"
)

func LogErrorPanic(action *action.Action, err error) {
	slog.Error(action.Name, "error", err)
	panic(err)
}

func LogErrorPrintStderrPanic(action *action.Action, errMsg string, stackTrace string) {
	slog.Error(action.Name, "error", errMsg)
	fmt.Println("Stderr: ", stackTrace)
	panic(errors.New(errMsg))
}

func LogDebug(action *action.Action, err error) {
	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		slog.Warn(action.Name, "text", err)
	}
}
