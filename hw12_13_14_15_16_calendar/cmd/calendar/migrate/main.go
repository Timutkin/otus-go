//go:build migrations
// +build migrations

package main

import (
	"context"
	"flag"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog/log"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/config"
	_ "github.com/timutkin/otus-go/hw12_13_14_15_calendar/migrations"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./configs/calendar_config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	cfg := config.NewCalendarConfig(configFile)

	if !cfg.DB.InMemory {
		db, err := goose.OpenDBWithDriver("pgx", cfg.DB.CollectDsn())
		if err != nil {
			log.Fatal().Err(err).Msg("error while connect to db")
		}
		defer db.Close()

		ctx := context.Background()
		if err = goose.RunContext(ctx, "up", db, "."); err != nil {
			log.Fatal().Err(err).Msg("goose run failed")
		}
	}
}
