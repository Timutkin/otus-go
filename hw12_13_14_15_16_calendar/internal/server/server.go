package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog/log"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/grpc/pb"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

type Server struct {
	sever        *http.Server
	grpcServer   *grpc.Server
	lg           Logger
	httpEndpoint string
	grpcEndpoint string
}

type Logger interface {
	InfoWithParams(msg string, params map[string]string)
	Fatal(msg string, err error)
}

type App struct {
	eventService pb.EventServiceServer
}

func NewApp(eventService pb.EventServiceServer) *App {
	return &App{eventService: eventService}
}

func NewServer(app *App, lg Logger, cfg config.Server) *Server {
	httpEndpoint := fmt.Sprintf("%s:%d", cfg.HTTPHost, cfg.HTTPPort)
	grpcEndpoint := fmt.Sprintf("%s:%d", cfg.GRPCHost, cfg.GRPCPort)
	server, err := createHTTPServer(grpcEndpoint, httpEndpoint, lg)
	if err != nil {
		lg.Fatal("creating http server error", err)
	}
	grpcServer := createGRPCServer(app, lg)
	return &Server{
		sever:        server,
		grpcServer:   grpcServer,
		httpEndpoint: httpEndpoint,
		grpcEndpoint: grpcEndpoint,
		lg:           lg,
	}
}

func (s *Server) Start(ctx context.Context) {
	go func() {
		if err := s.sever.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.lg.Fatal("failed to listen http "+s.sever.Addr, err)
		}
	}()
	go func() {
		lis, err := net.Listen("tcp", s.grpcEndpoint)
		if err != nil {
			s.lg.Fatal("failed to listen grpc "+s.grpcEndpoint, err)
		}
		err = s.grpcServer.Serve(lis)
		if err != nil {
			s.lg.Fatal("failed to serve grpc server ", err)
		}
	}()
	log.Info().Msg("calendar is running...")
	<-ctx.Done()
}

func createHTTPServer(grpcServerEndpoint, httpServerEndpoint string, lg Logger) (*http.Server, error) {
	jsonMarshaler := &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}

	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, jsonMarshaler),
	)

	noCredentials := grpc.WithTransportCredentials(insecure.NewCredentials())
	opts := []grpc.DialOption{noCredentials}
	err := pb.RegisterEventServiceHandlerFromEndpoint(context.Background(), mux, grpcServerEndpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("register event service handler : %w", err)
	}

	srv := &http.Server{
		Addr:              httpServerEndpoint,
		Handler:           httpLoggingMiddleware(mux, lg),
		ReadHeaderTimeout: time.Second * 10,
	}
	return srv, nil
}

func createGRPCServer(app *App, lg Logger) *grpc.Server {
	grpcLoggingInterceptor := NewGrpcLoggingInterceptor(lg)
	s := grpc.NewServer(grpc.UnaryInterceptor(grpcLoggingInterceptor.grpcLoggingMiddleware))
	pb.RegisterEventServiceServer(s, app.eventService)
	return s
}

func (s *Server) Stop(ctx context.Context) error {
	err := s.sever.Shutdown(ctx)
	s.grpcServer.GracefulStop()
	return err
}
