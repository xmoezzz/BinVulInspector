package log

import (
	"errors"
	"io"
	"os"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zapio"
)

const (
	DebugLevel = "debug"
)

type Logger interface {
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})
	Info(args ...interface{})
	Infof(template string, args ...interface{})
	Warn(args ...interface{})
	Warnf(template string, args ...interface{})
	Error(args ...interface{})
	Errorf(template string, args ...interface{})
}

var (
	ErrUnexpectedLogLevel = errors.New("unexpected log level")
)

func NewFileWriter(level string, filepath string) (io.Writer, error) {
	var zapLevel zapcore.Level
	{
		var err error
		zapLevel, err = convertZapLevel(level)
		if err != nil {
			return nil, err
		}
	}
	logger, err := newFileLogger(zapLevel, filepath)
	if err != nil {
		return nil, err
	}
	return &zapio.Writer{Log: logger, Level: zapLevel}, nil
}

func NewConsoleLogger(level string) (Logger, error) {
	var zapLevel zapcore.Level
	{
		var err error
		zapLevel, err = convertZapLevel(level)
		if err != nil {
			return nil, err
		}
	}

	// 日志格式
	var encoderConfig zapcore.EncoderConfig
	{
		encoderConfig = defaultZapEncoderConfig()
	}

	var consoleWriteSyncer zapcore.WriteSyncer
	{
		// High-priority output should also go to standard error, and low-priority
		// output should also go to standard out.
		consoleWriteSyncer = zapcore.Lock(os.Stdout)
	}

	var zapLogCore zapcore.Core
	{
		zapLogCore = zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), consoleWriteSyncer, zapLevel)
	}

	var zapLogOpts []zap.Option
	{
		zapLogOpts = defaultZapLogOpts(zapLevel)
	}

	logger := zap.New(zapLogCore, zapLogOpts...)

	logger.WithOptions()

	return logger.Sugar(), nil
}

func newFileLogger(level zapcore.Level, filepath string) (*zap.Logger, error) {

	// 日志格式
	var encoderConfig zapcore.EncoderConfig
	{
		encoderConfig = defaultZapEncoderConfig()
	}

	var fileWriteSyncer zapcore.WriteSyncer
	{
		hook := lumberjack.Logger{
			Filename:   filepath,
			MaxSize:    128,
			MaxAge:     7,
			MaxBackups: 30,
			Compress:   false,
		}
		writes := []zapcore.WriteSyncer{zapcore.AddSync(&hook)}
		fileWriteSyncer = zapcore.NewMultiWriteSyncer(writes...)
	}

	var zapLogCore zapcore.Core
	{
		zapLogCore = zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), fileWriteSyncer, level)
	}

	var zapLogOpts []zap.Option
	{
		zapLogOpts = defaultZapLogOpts(level)
	}

	logger := zap.New(zapLogCore, zapLogOpts...)

	return logger, nil
}

func defaultZapLogOpts(zapLevel zapcore.Level) (zapLogOpts []zap.Option) {
	zapLogOpts = append(zapLogOpts, zap.ErrorOutput(zapcore.Lock(os.Stderr)))
	// 日志设置为开发模式
	if zapLevel == zap.DebugLevel {
		zapLogOpts = append(zapLogOpts, zap.Development())
	}
	// 添加调用信息
	zapLogOpts = append(zapLogOpts, zap.AddCallerSkip(1))
	return zapLogOpts
}

func defaultZapEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "Time",
		LevelKey:       "Level",
		NameKey:        "Name",
		CallerKey:      "Caller",
		MessageKey:     "Msg",
		StacktraceKey:  "St",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// convert string to zapcore.Level
func convertZapLevel(level string) (zapLevel zapcore.Level, err error) {
	switch level {
	case DebugLevel:
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		err = ErrUnexpectedLogLevel
	}
	return zapLevel, err
}
