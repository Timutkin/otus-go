package config

import (
	"context"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/file"
	"github.com/rs/zerolog/log"
)

type CalendarSchedulerConfig struct {
	Rabbit Rabbit                        `config:"rabbit"`
	Logger CalendarSchedulerLoggerConfig `config:"logging"`
	DB     DBConf                        `config:"db"`
}

type Rabbit struct {
	ConnectionString string `yaml:"connection-string"` //nolint:tagliatelle
	QueueName        string `yaml:"queue"`
}

type CalendarSchedulerLoggerConfig struct {
	Level string `config:"level,required"`
}

func NewSchedulerConfig(pathToYaml string) CalendarSchedulerConfig {
	cfg := CalendarSchedulerConfig{
		Logger: CalendarSchedulerLoggerConfig{
			Level: "info",
		},
		DB: DBConf{
			InMemory: false,
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			Dbname:   "postgres",
		},
	}
	loader := confita.NewLoader(file.NewBackend(pathToYaml))
	if err := loader.Load(context.Background(), &cfg); err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}
	return cfg
}
