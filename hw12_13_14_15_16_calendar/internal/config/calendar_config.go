package config

import (
	"context"
	"fmt"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/file"
	"github.com/rs/zerolog/log"
)

type CalendarConfig struct {
	Logger LoggerConf `config:"logging"`
	DB     DBConf     `config:"db"`
	Server Server     `config:"server"`
}

type LoggerConf struct {
	Level string `config:"level,required"`
}

type Server struct {
	HTTPHost string `yaml:"http-host"` //nolint:tagliatelle
	HTTPPort int    `yaml:"http-port"` //nolint:tagliatelle
	GRPCHost string `yaml:"grpc-host"` //nolint:tagliatelle
	GRPCPort int    `yaml:"grpc-port"` //nolint:tagliatelle
}

type DBConf struct {
	InMemory bool     `yaml:"in-memory"` //nolint:tagliatelle
	Host     string   `config:"host"`
	Port     int      `config:"port"`
	User     string   `config:"user"`
	Password string   `config:"password"`
	Dbname   string   `config:"dbname"`
	Tables   DBTables `config:"tables"`
}

type DBTables struct {
	Schema string `config:"schema"`
}

func (c DBConf) CollectDsn() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.Dbname)
}

func NewCalendarConfig(pathToYaml string) CalendarConfig {
	cfg := CalendarConfig{
		Logger: LoggerConf{
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
		Server: Server{
			HTTPHost: "localhost",
			HTTPPort: 8080,
			GRPCHost: "localhost",
			GRPCPort: 50051,
		},
	}
	loader := confita.NewLoader(file.NewBackend(pathToYaml))
	if err := loader.Load(context.Background(), &cfg); err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}
	return cfg
}
