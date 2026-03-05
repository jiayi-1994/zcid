package logging

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

var runtimeLevel = new(slog.LevelVar)

func Init(level string) {
	runtimeLevel.Set(parseLevel(level))
	handler := NewMaskingHandler(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: runtimeLevel}))
	slog.SetDefault(slog.New(handler))
}

func SetLevel(level string) error {
	lvl, err := parseLevelStrict(level)
	if err != nil {
		return err
	}
	runtimeLevel.Set(lvl)
	return nil
}

func CurrentLevel() string {
	switch runtimeLevel.Level() {
	case slog.LevelDebug:
		return "debug"
	case slog.LevelInfo:
		return "info"
	case slog.LevelWarn:
		return "warn"
	case slog.LevelError:
		return "error"
	default:
		return runtimeLevel.Level().String()
	}
}

func parseLevel(level string) slog.Level {
	lvl, err := parseLevelStrict(level)
	if err != nil {
		return slog.LevelInfo
	}
	return lvl
}

func parseLevelStrict(level string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "", "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("invalid log level: %s", level)
	}
}
