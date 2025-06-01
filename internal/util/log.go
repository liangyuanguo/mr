package utility

import (
	"log/slog"
	"os"
)

var LoggerDict = make(map[string]*slog.Logger)

func GetDefaultLogger() *slog.Logger {
	v, exit := LoggerDict["default"]
	if !exit {
		v = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo, // 设置日志级别为 Info
		}))
		LoggerDict["default"] = v
		slog.SetDefault(v)
	}
	return v
}
