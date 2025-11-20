package internalhttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/config"
)

type Server struct {
	sever *http.Server
	lg    Logger
	app   Application
}

type Logger interface {
	InfoWithParams(msg string, params map[string]string)
}

type Application interface { // TODO
}

func NewServer(lg Logger, app Application, cfg config.Server) *Server {
	gin.SetMode(cfg.Mode)
	ginApp := gin.New()
	ginApp.Use(loggingMiddleware(lg))
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	return &Server{
		sever: &http.Server{
			Addr:              addr,
			Handler:           ginApp,
			ReadHeaderTimeout: time.Second * 30,
		},
		lg:  lg,
		app: app,
	}
}

func (s *Server) Start(ctx context.Context) {
	go func() {
		if err := s.sever.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("failed to listen " + s.sever.Addr)
		}
	}()
	log.Info().Msg("calendar is running...")
	<-ctx.Done()
}

func (s *Server) Stop(ctx context.Context) error {
	err := s.sever.Shutdown(ctx)
	return err
}
