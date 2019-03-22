package slog

import (
	"go.uber.org/zap"
)

type zapLoggerWrapper struct {
	log *zap.SugaredLogger
}

func NewZapLoggerWrapper(log *zap.Logger) Logger {
	sugar := log.WithOptions(zap.AddCallerSkip(2)).Sugar() // 跳过2层, 直接定位到surgemq打日志的源码
	sugar = sugar.Named("surgemq")
	return &zapLoggerWrapper{log: sugar}
}

func (l *zapLoggerWrapper) Debug(args ...interface{}) {
	l.log.Debug(args)
}

func (l *zapLoggerWrapper) Debugln(args ...interface{}) {
	l.log.Debug(args)
}

func (l *zapLoggerWrapper) Debugf(format string, args ...interface{}) {
	l.log.Debugf(format, args...)
}

func (l *zapLoggerWrapper) Info(args ...interface{}) {
	l.log.Info(args)
}

func (l *zapLoggerWrapper) Infoln(args ...interface{}) {
	l.log.Info(args)
}

func (l *zapLoggerWrapper) Infof(format string, args ...interface{}) {
	l.log.Infof(format, args...)
}

func (l *zapLoggerWrapper) Warning(args ...interface{}) {
	l.log.Warn(args)
}

func (l *zapLoggerWrapper) Warningln(args ...interface{}) {
	l.log.Warn(args)
}

func (l *zapLoggerWrapper) Warningf(format string, args ...interface{}) {
	l.log.Warnf(format, args...)
}

func (l *zapLoggerWrapper) Error(args ...interface{}) {
	l.log.Error(args)
}

func (l *zapLoggerWrapper) Errorln(args ...interface{}) {
	l.log.Error(args)
}

func (l *zapLoggerWrapper) Errorf(format string, args ...interface{}) {
	l.log.Errorf(format, args...)
}

func (l *zapLoggerWrapper) Fatal(args ...interface{}) {
	l.log.Fatal(args)
}

func (l *zapLoggerWrapper) Fatalln(args ...interface{}) {
	l.log.Fatal(args)
}

func (l *zapLoggerWrapper) Fatalf(format string, args ...interface{}) {
	l.log.Fatalf(format, args...)
}

func (l *zapLoggerWrapper) V(level int) bool {
	return true
}
