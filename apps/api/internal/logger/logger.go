package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func New(appEnv string) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339

	if appEnv == "development" {
		return log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).With().Str("service", "forgeflow-api").Logger()
	}

	return zerolog.New(os.Stdout).With().
		Timestamp().
		Str("service", "forgeflow-api").
		Logger()
}
