package l

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Init initializes the logger.
func createLogger() *zap.Logger {
	stdout := zapcore.AddSync(os.Stdout)

	level := zap.NewAtomicLevelAt(zap.DebugLevel)

	// productionCfg := zap.NewProductionEncoderConfig()
	// productionCfg.TimeKey = "timestamp"
	// productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	developmentCfg := zap.NewDevelopmentEncoderConfig()
	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, stdout, level),
	)

	return zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)
}

var logger = createLogger()

func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	logger.Panic(msg, fields...)
}

func InfoF(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logger.Info(msg)
}

func WarnF(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logger.Warn(msg)
}

func ErrorF(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logger.Error(msg)
}

func FatalF(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logger.Fatal(msg)
}

func DebugF(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logger.Debug(msg)
}

func PanicF(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logger.Panic(msg)
}
