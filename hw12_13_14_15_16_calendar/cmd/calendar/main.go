package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/app"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/config"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/logger"
	internalhttp "github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/server/http"
	memorystorage "github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/storage/sql"
	_ "github.com/timutkin/otus-go/hw12_13_14_15_calendar/migrations"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./configs/config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	cfg := config.NewConfig(configFile)
	level, err := zerolog.ParseLevel(cfg.Logger.Level)
	if err != nil {
		log.Fatal().Err(err).Msg("error while parsing log level")
	}
	zerolog.SetGlobalLevel(level)
	logg := logger.New()

	var storage app.Storage
	if cfg.DB.InMemory {
		logg.Info("work with in-memory mod ...")
		storage = memorystorage.New()
	} else {
		logg.Info("work with postgresql...")
		storage = sqlstorage.New(cfg.DB)
	}

	calendar := app.New(logg, storage)

	server := internalhttp.NewServer(logg, calendar, cfg.Server)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server", err)
		}
	}()

	server.Start(ctx)
}
