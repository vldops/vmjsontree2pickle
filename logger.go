package main

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func makeLogger() {
	cfg := zap.Config{
		Encoding:         "json",
		Development:      true,
		Level:            zap.NewAtomicLevelAt(config.Logger.LevelEncoded),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{

			MessageKey:       "message",
			ConsoleSeparator: " ",
			LevelKey:         "level",
			EncodeLevel:      zapcore.CapitalLevelEncoder,
			TimeKey:          "time",
			EncodeTime:       zapcore.RFC3339TimeEncoder,
			CallerKey:        "caller",
			EncodeCaller:     zapcore.ShortCallerEncoder,
		},
		InitialFields: map[string]interface{}{
			"appName": "app",
		},
	}

	logger, err = cfg.Build()
	if err != nil {
		log.Fatalln(err)
	}

}
