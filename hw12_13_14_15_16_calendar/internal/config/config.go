package config

import (
	"context"
	"fmt"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/file"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Logger LoggerConf `config:"logging"`
	DB     DBConf     `config:"db"`
	Server Server     `config:"server"`
}

type LoggerConf struct {
	Level string `config:"level,required"`
}

type Server struct {
	Host string `config:"host"`
	Port int    `config:"port"`
	Mode string `config:"mode"`
}

type DBConf struct {
	InMemory bool     `yaml:"in-memory"` //nolint:tagliatelle
	Host     string   `config:"host"`
	Port     int      `config:"port"`
	User     string   `config:"user"`
	Password string   `config:"password"`
	Dbname   string   `config:"dbname"`
	DBTables DBTables `config:"tables"`
}
type DBTables struct {
	Events string `config:"events"`
	Schema string `config:"schema"`
}

func (c DBConf) CollectDsn() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.Dbname)
}

func NewConfig(pathToYaml string) Config {
	cfg := Config{
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
			Host: "localhost",
			Port: 8080,
			Mode: "release",
		},
	}
	loader := confita.NewLoader(file.NewBackend(pathToYaml))
	if err := loader.Load(context.Background(), &cfg); err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}
	return cfg
}
