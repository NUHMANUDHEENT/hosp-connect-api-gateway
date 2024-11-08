package logs

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Setup logrus and lumbejack for logging
func NewLogger() *logrus.Logger {
	logger := logrus.New()

	// Set up Lumberjack logger
	lumberjackLogger := &lumberjack.Logger{
		Filename:   "logs/api_gateway.log",
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30, 
		Compress:   true,
	}

	multiWriter := io.MultiWriter(os.Stdout, lumberjackLogger)
	logger.SetOutput(multiWriter)
	logger.SetFormatter(&logrus.JSONFormatter{})

	return logger
}
