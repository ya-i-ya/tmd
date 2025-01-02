package logger

import (
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func SetupLogger(logPath string, level string) error {
	levelMap := map[string]zerolog.Level{
		"debug": zerolog.DebugLevel,
		"info":  zerolog.InfoLevel,
		"warn":  zerolog.WarnLevel,
		"error": zerolog.ErrorLevel,
		"fatal": zerolog.FatalLevel,
		"panic": zerolog.PanicLevel,
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if lvl, exists := levelMap[level]; exists {
		zerolog.SetGlobalLevel(lvl)
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out: &lumberjack.Logger{
			Filename:   logPath,
			MaxSize:    10, // megabytes
			MaxAge:     28, // days
			MaxBackups: 5,
			Compress:   true,
		},
		TimeFormat: time.RFC3339,
	})

	return nil
}
