package logger

import (
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func SetupLogger(logPath string) error {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out: &lumberjack.Logger{
			Filename:   logPath,
			MaxSize:    10,
			MaxAge:     28,
			MaxBackups: 5,
			Compress:   true,
		},
		TimeFormat: time.RFC3339,
	})

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	return nil
}
