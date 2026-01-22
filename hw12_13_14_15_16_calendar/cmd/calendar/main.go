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
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/config"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/logger"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/mapper"
	internalhttp "github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/server"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/service"
	memorystorage "github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./configs/calendar_config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	cfg := config.NewCalendarConfig(configFile)
	level, err := zerolog.ParseLevel(cfg.Logger.Level)
	if err != nil {
		log.Fatal().Err(err).Msg("error while parsing log level")
	}
	zerolog.SetGlobalLevel(level)
	logg := logger.New()

	var storage service.Storage
	if cfg.DB.InMemory {
		logg.Info("work with in-memory mod ...")
		storage = memorystorage.New()
	} else {
		logg.Info("work with postgresql...")
		sql := sqlstorage.New(cfg.DB)
		err := sql.Connect(context.Background())
		if err != nil {
			logg.Fatal("failed connect to db", err)
		}
		storage = sql
	}

	eventService := service.NewEventService(storage, logg, mapper.EventMapper{})
	app := internalhttp.NewApp(eventService)
	server := internalhttp.NewServer(app, logg, cfg.Server)

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
