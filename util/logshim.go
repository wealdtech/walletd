package util

import (
	"fmt"

	"github.com/rs/zerolog"
)

// LogShim shims a zerolog logger to provide a traditional logger interface.
type LogShim struct {
	log zerolog.Logger
}

func NewLogShim(log zerolog.Logger) LogShim {
	return LogShim{log: log}
}

func (l LogShim) Fatal(args ...interface{}) {
	l.log.Fatal().Msg(fmt.Sprint(args...))
}

func (l LogShim) Fatalf(format string, args ...interface{}) {
	l.log.Fatal().Msgf(format, args...)
}

func (l LogShim) Fatalln(args ...interface{}) {
	l.Fatal(args...)
}

func (l LogShim) Error(args ...interface{}) {
	l.log.Error().Msg(fmt.Sprint(args...))
}

func (l LogShim) Errorf(format string, args ...interface{}) {
	l.log.Error().Msgf(format, args...)
}

func (l LogShim) Errorln(args ...interface{}) {
	l.Error(args...)
}

func (l LogShim) Warning(args ...interface{}) {
	l.log.Warn().Msg(fmt.Sprint(args...))
}

func (l LogShim) Warningf(format string, args ...interface{}) {
	l.log.Warn().Msgf(format, args...)
}

func (l LogShim) Warningln(args ...interface{}) {
	l.Warning(args...)
}

func (l LogShim) Info(args ...interface{}) {
	l.log.Info().Msg(fmt.Sprint(args...))
}

func (l LogShim) Infof(format string, args ...interface{}) {
	l.log.Info().Msgf(format, args...)
}

func (l LogShim) Infoln(args ...interface{}) {
	l.Info(args...)
}

func (l LogShim) Debug(args ...interface{}) {
	l.log.Debug().Msg(fmt.Sprint(args...))
}

func (l LogShim) Debugf(format string, args ...interface{}) {
	l.log.Debug().Msgf(format, args...)
}

func (l LogShim) Debugln(args ...interface{}) {
	l.Debug(args...)
}

func (l LogShim) Print(args ...interface{}) {
	l.log.Info().Msg(fmt.Sprint(args...))
}

func (l LogShim) Printf(format string, args ...interface{}) {
	l.log.Info().Msgf(format, args...)
}

func (l LogShim) Println(args ...interface{}) {
	l.Print(args...)
}

func (l LogShim) V(level int) bool {
	return true
}
