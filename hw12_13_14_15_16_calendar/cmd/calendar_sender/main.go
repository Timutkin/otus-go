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
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/sender"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./configs/calendar_sender_config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	cfg := config.NewSenderConfig(configFile)

	level, err := zerolog.ParseLevel(cfg.Logger.Level)
	if err != nil {
		log.Fatal().Err(err).Msg("parsing logging level")
	}
	zerolog.SetGlobalLevel(level)
	logg := logger.New()

	rabbitClient := client.NewRabbitClient(cfg.Rabbit.ConnectionString, logg)
	notificationSender := sender.NewNotificationSender(rabbitClient, cfg.Rabbit.QueueName, logg)
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()
	notificationSender.StartListening(ctx)
}
