package logger

import (
	"os"

	"github.com/rs/zerolog"
)

var lg = zerolog.New(os.Stdout).With().Timestamp().Logger()

type Logger struct {
	Lg zerolog.Logger
}

func New() *Logger {
	return &Logger{Lg: lg}
}

func (l *Logger) InfoWithParams(msg string, params map[string]string) {
	lg := l.Lg.Info()
	addParams(msg, params, lg)
}

func (l *Logger) ErrorWithParams(msg string, params map[string]string, err error) {
	lg := l.Lg.Error().Err(err)
	addParams(msg, params, lg)
}

func (l *Logger) DebugWithParams(msg string, params map[string]string) {
	lg := l.Lg.Debug()
	addParams(msg, params, lg)
}

func (l *Logger) ErrorWithAny(msg string, name string, param any) {
	l.Lg.Error().Any(name, param).Msg(msg)
}

func addParams(msg string, params map[string]string, lg *zerolog.Event) {
	for k, v := range params {
		lg = lg.Any(k, v)
	}
	lg.Msg(msg)
}

func (l *Logger) Error(msg string, err error) {
	l.Lg.Error().Err(err).Msg(msg)
}

func (l *Logger) Fatal(msg string, err error) {
	l.Lg.Fatal().Err(err).Msg(msg)
}

func (l *Logger) Info(msg string) {
	l.Lg.Info().Msg(msg)
}

func (l *Logger) Debug(msg string) {
	l.Lg.Info().Msg(msg)
}
