package logger

import (
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

func SetupLogger(logPath string) error {
	logrus.SetOutput(&lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    10,
		MaxAge:     28,
		MaxBackups: 5,
		Compress:   true,
	})
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.InfoLevel)
	return nil
}
