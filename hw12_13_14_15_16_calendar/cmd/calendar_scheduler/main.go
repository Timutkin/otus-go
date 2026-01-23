package main

import (
	"context"
	"flag"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/client"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/config"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/logger"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/scheduler"
	sqlstorage "github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./configs/calendar_scheduler_config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	cfg := config.NewSchedulerConfig(configFile)

	level, err := zerolog.ParseLevel(cfg.Logger.Level)
	if err != nil {
		log.Fatal().Err(err).Msg("parsing jobs")
	}
	zerolog.SetGlobalLevel(level)
	logg := logger.New()

	rabbitClient := client.NewRabbitClient(cfg.Rabbit.ConnectionString, logg)
	s := scheduler.NewScheduler(logg)

	sql := sqlstorage.New(cfg.DB)
	err = sql.Connect(context.Background())
	if err != nil {
		logg.Fatal("failed connect to db", err)
	}

	notificationScheduler := scheduler.NewNotificationScheduler(sql, rabbitClient, logg, cfg.Rabbit.QueueName)
	err = s.CreateJobs(notificationScheduler.GetJobs())
	if err != nil {
		log.Fatal().Err(err).Msg("create jobs")
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()
		s.Shutdown()
	}()

	s.Start(ctx)
}
