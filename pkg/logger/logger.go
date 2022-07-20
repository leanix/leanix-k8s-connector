package logger

import (
	"bytes"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var zapLog *zap.Logger

func Init() bytes.Buffer {
	logFile := new(bytes.Buffer)
	zapLog = zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.Lock(zapcore.AddSync(logFile)), zapcore.InfoLevel,
		))
	zapLog = zapLog.WithOptions(zap.WrapCore(
		func(c zapcore.Core) zapcore.Core {
			ucEncoderCfg := encoderCfg
			ucEncoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
			return zapcore.NewTee(
				c,
				zapcore.NewCore(
					zapcore.NewConsoleEncoder(ucEncoderCfg),
					zapcore.Lock(os.Stdout),
					zapcore.InfoLevel,
				))
		}))

	return *logFile
}

func Sync() error {
	err := zapLog.Sync()
	if err != nil {
		return err
	}
	return nil
}

func Info(message string) {
	zapLog.Info(message)
}

func Infof(format string, arguments ...interface{}) {
	zapLog.Info(fmt.Sprintf(format, arguments))
}

func Debug(message string, fields ...zap.Field) {
	zapLog.Debug(message, fields...)
}

func Debugf(message string, arguments ...interface{}) {
	zapLog.Debug(fmt.Sprintf(message, arguments))
}

func Error(message string, err error) {
	zapLog.Error(message, zap.NamedError("error", err))
}

func Errorf(message string, arguments ...interface{}) {
	zapLog.Error(fmt.Sprintf(message, arguments))
}

func Fatal(message string, fields ...zap.Field) {
	zapLog.Fatal(message, fields...)
}

var encoderCfg = zapcore.EncoderConfig{MessageKey: "msg",
	NameKey:      "name",
	LevelKey:     "level",
	EncodeLevel:  zapcore.LowercaseLevelEncoder,
	CallerKey:    "caller",
	EncodeCaller: zapcore.ShortCallerEncoder,
	// TimeKey: "time",
	// EncodeTime: zapcore.ISO8601TimeEncoder,
}
