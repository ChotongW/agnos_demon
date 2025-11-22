package helpers

import (
	"os"

	"github.com/sirupsen/logrus"
)

func CreateLogger(level string, output *os.File, hooks []logrus.Hook, format string) (*logrus.Logger, error) {
	logger := logrus.New()

	if level != "" {
		lvl, err := logrus.ParseLevel(level)
		if err != nil {
			return nil, err
		}
		logger.SetLevel(lvl)
	} else {
		logger.SetLevel(logrus.DebugLevel)
	}

	if output != nil {
		logger.SetOutput(output)
	} else {
		logger.SetOutput(os.Stdout)
	}

	if format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			ForceColors: true,
		})
	}

	// Add hooks if any
	for _, hook := range hooks {
		logger.AddHook(hook)
	}

	return logger, nil
}
