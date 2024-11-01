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
		Filename:   "/home/nuhmanudheen-t/Broto/2ndProject/HospitalConnect/api_gateway/logs/api_gateway.log",
		MaxSize:    10, // MB
		MaxBackups: 5,
		MaxAge:     30, // days
		Compress:   true,
	}

	multiWriter := io.MultiWriter(os.Stdout, lumberjackLogger)
	logger.SetOutput(multiWriter)
	logger.SetFormatter(&logrus.JSONFormatter{})

	return logger
}
