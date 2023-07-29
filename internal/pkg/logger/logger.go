package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/wyattjychen/hades/internal/pkg/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var defaultLogger *zap.Logger

func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func Init(projectName string, level string, format, prefix, director string, showLine bool, encodeLevel string, stacktraceKey string, logInConsole bool) (logger *zap.Logger) {
	if ok := Exists(fmt.Sprintf("%s/%s", projectName, director)); !ok {
		fmt.Printf("create %v directory to save log.\n", director)
		_ = os.Mkdir(fmt.Sprintf("%s/%s", projectName, director), os.ModePerm)
	}
	debugPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.DebugLevel
	})
	infoPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.InfoLevel
	})
	warnPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.WarnLevel
	})
	errorPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev >= zap.ErrorLevel
	})
	cores := make([]zapcore.Core, 0)
	switch level {
	case "info":
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_info.log", projectName, director), infoPriority))
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_warn.log", projectName, director), warnPriority))
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_error.log", projectName, director), errorPriority))
	case "warn":
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_warn.log", projectName, director), warnPriority))
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_error.log", projectName, director), errorPriority))
	case "error":
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_error.log", projectName, director), errorPriority))
	default:
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_debug.log", projectName, director), debugPriority))
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_info.log", projectName, director), infoPriority))
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_warn.log", projectName, director), warnPriority))
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_error.log", projectName, director), errorPriority))
	}
	logger = zap.New(zapcore.NewTee(cores[:]...), zap.AddCaller())

	if showLine {
		logger = logger.WithOptions(zap.AddCaller())
	}
	defaultLogger = logger
	return logger
}

func getEncoderConfig(prefix, encodeLevel, stacktraceKey string) (config zapcore.EncoderConfig) {
	config = zapcore.EncoderConfig{
		MessageKey:    "message",
		LevelKey:      "level",
		TimeKey:       "time",
		NameKey:       "logger",
		CallerKey:     "caller",
		StacktraceKey: stacktraceKey,
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format(prefix + utils.TimeFormatDateV4))
		},
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}
	switch {
	case encodeLevel == "LowercaseLevelEncoder":
		config.EncodeLevel = zapcore.LowercaseLevelEncoder
	case encodeLevel == "LowercaseColorLevelEncoder":
		config.EncodeLevel = zapcore.LowercaseColorLevelEncoder
	case encodeLevel == "CapitalLevelEncoder":
		config.EncodeLevel = zapcore.CapitalLevelEncoder
	case encodeLevel == "CapitalColorLevelEncoder":
		config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	default:
		config.EncodeLevel = zapcore.LowercaseLevelEncoder
	}
	return config
}

func getEncoderCore(logInConsole bool, prefix, format, encodeLevel, stacktraceKey string, fileName string, level zapcore.LevelEnabler) (core zapcore.Core) {
	writer := getWriteSyncer(logInConsole, fileName) // 使用file-rotatelogs进行日志分割
	return zapcore.NewCore(getEncoder(prefix, format, encodeLevel, stacktraceKey), writer, level)
}

func getEncoder(prefix, format, encodeLevel, stacktraceKey string) zapcore.Encoder {
	if format == "json" {
		return zapcore.NewJSONEncoder(getEncoderConfig(prefix, encodeLevel, stacktraceKey))
	}
	return zapcore.NewConsoleEncoder(getEncoderConfig(prefix, encodeLevel, stacktraceKey))
}

func getWriteSyncer(logInConsole bool, file string) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   file,
		MaxSize:    10,
		MaxBackups: 200,
		MaxAge:     30,
		Compress:   true,
	}

	if logInConsole {
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(lumberJackLogger))
	}
	return zapcore.AddSync(lumberJackLogger)
}

func Shutdown() {
	defaultLogger.Sync()
}
func GetLogger() *zap.Logger {
	return defaultLogger
}
