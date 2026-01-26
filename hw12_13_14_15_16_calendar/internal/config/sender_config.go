package config

import (
	"context"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/file"
	"github.com/rs/zerolog/log"
)

type CalendarSenderConfig struct {
	Rabbit Rabbit                     `config:"rabbit"`
	DB     DBConf                     `config:"db"`
	Logger CalendarSenderLoggerConfig `config:"logging"`
}

type CalendarSenderLoggerConfig struct {
	Level string `config:"level,required"`
}

func NewSenderConfig(pathToYaml string) CalendarSenderConfig {
	cfg := CalendarSenderConfig{
		Logger: CalendarSenderLoggerConfig{
			Level: "info",
		},
	}
	loader := confita.NewLoader(file.NewBackend(pathToYaml))
	if err := loader.Load(context.Background(), &cfg); err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}
	return cfg
}
