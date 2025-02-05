package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogLevel string

func (l LogLevel) String() string {
	return string(l)
}

const (
	DEBUG   LogLevel = "debug"
	INFO    LogLevel = "info"
	WARNING LogLevel = "warning"
	ERROR   LogLevel = "error"
	FATAL   LogLevel = "fatal"
	PANIC   LogLevel = "panic"
)

// GetAllLevels return all log levels. Used in validation.
func GetAllLevels() []interface{} {
	return []interface{}{
		DEBUG.String(), INFO.String(), WARNING.String(), ERROR.String(), FATAL.String(), PANIC.String(),
	}
}

type LogFormat string

const (
	FormatConsole LogFormat = "console"
	FormatJSON    LogFormat = "json"
)

type SugaredLogger = zap.SugaredLogger

type logger struct {
	*SugaredLogger
}

// TODO я знаю что тут куча всего чеего еще не используется,
// TODO но это просто сокпированый логгер уже с проекта моего старого).

// Logger common interface.
type Logger interface {
	Debug(...interface{})
	Debugf(string, ...interface{})
	Info(...interface{})
	Infof(string, ...interface{})
	Infoln(...interface{})
	Infow(string, ...interface{})
	Warn(...interface{})
	Warnf(string, ...interface{})
	Error(...interface{})
	Errorf(string, ...interface{})
	Fatal(...interface{})
	Fatalf(string, ...interface{})
	Panic(...interface{})
	Panicf(string, ...interface{})
	With(...interface{}) Logger
	Sync() error
	Sugar() *SugaredLogger
}

func initLogger(level LogLevel, consoleColored bool, timeKey string) *zap.Logger {
	atom := zap.NewAtomicLevel()

	encoderCfg := zap.NewProductionEncoderConfig()

	encoderCfg.TimeKey = "ts"
	if timeKey != "" {
		encoderCfg.TimeKey = timeKey
	}

	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	if consoleColored {
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	encoder := zapcore.NewConsoleEncoder(encoderCfg)

	logger := zap.New(zapcore.NewCore(
		encoder,
		zapcore.Lock(os.Stdout),
		atom),
		zap.AddCaller(),
	)

	switch level {
	case DEBUG:
		atom.SetLevel(zap.DebugLevel)
	case INFO:
		atom.SetLevel(zap.InfoLevel)
	case WARNING:
		atom.SetLevel(zap.WarnLevel)
	case ERROR:
		atom.SetLevel(zap.ErrorLevel)
	case FATAL:
		atom.SetLevel(zap.FatalLevel)
	case PANIC:
		atom.SetLevel(zap.PanicLevel)
	default:
		atom.SetLevel(zap.InfoLevel)
	}

	return logger
}

// New - init new logger with options.
func New(opts ...Option) Logger {
	options := Options{}
	for _, opt := range opts {
		opt(&options)
	}

	if options.LogLevel == "" {
		options.LogLevel = DEBUG
	}

	l := initLogger(
		options.LogLevel,
		options.ConsoleColored,
		options.TimeKey,
	)

	if options.AppName != "" {
		l = l.With(
			zap.String("app", options.AppName),
		)
	}

	if options.AppVersion != "" {
		l = l.With(
			zap.String("version", options.AppVersion),
		)
	}

	return &logger{
		SugaredLogger: l.Sugar(),
	}
}

func (l *logger) Sugar() *SugaredLogger {
	return l.SugaredLogger
}

func (l logger) With(args ...any) Logger {
	return &logger{l.SugaredLogger.With(args...)}
}
