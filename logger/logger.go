package logger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func Init(level, output string) error {
	var w io.Writer

	switch output {
	case "stdout":
		w = os.Stdout
	case "stderr":
		w = os.Stderr
	default:
		f, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file %s: %w", output, err)
		}
		w = f
	}

	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	Log = zerolog.New(w).
		Level(lvl).
		With().
		Timestamp().
		Caller().
		Logger()

	zerolog.TimeFieldFormat = time.RFC3339
	return nil
}

func Debug(msg string) {
	Log.Debug().Msg(msg)
}

func Info(msg string) {
	Log.Info().Msg(msg)
}

func Warn(msg string) {
	Log.Warn().Msg(msg)
}

func Error(msg string, err error) {
	Log.Error().Err(err).Msg(msg)
}

func Fatal(msg string, err error) {
	Log.Fatal().Err(err).Msg(msg)
}