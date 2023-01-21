package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var zapLog *zap.Logger
var atom zap.AtomicLevel

func init() {

	atom = zap.NewAtomicLevel()
	cfg := zap.NewProductionConfig()
	cfg.Encoding = "console"
	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stderr"}
	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	cfg.Level = atom
	//cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	zapLog, _ = cfg.Build(zap.AddCallerSkip(1))

}

func SetLevel(l zapcore.Level) {
	atom.SetLevel(l)
}

func Error(msg string, fields ...zap.Field) {
	zapLog.Error(msg, fields...)
}

func Errorf(format string, fields ...any) {
	zapLog.Error(fmt.Sprintf(format, fields...))
}

func Info(msg string, fields ...zap.Field) {
	zapLog.Info(msg, fields...)
}

func Infof(format string, a ...any) {
	zapLog.Info(fmt.Sprintf(format, a...))
}

func Debug(msg string, fields ...zap.Field) {
	zapLog.Debug(msg, fields...)
}

func Debugf(format string, a ...any) {
	zapLog.Debug(fmt.Sprintf(format, a...))
}

func Fatal(msg string, fields ...zap.Field) {
	zapLog.Fatal(msg, fields...)
}

func Fatalf(format string, a ...any) {
	zapLog.Fatal(fmt.Sprintf(format, a...))
}

func Warn(msg string, fields ...zap.Field) {
	zapLog.Warn(msg, fields...)
}
