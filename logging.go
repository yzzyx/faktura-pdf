package main

import (
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/lestrrat-go/strftime"
	"github.com/yzzyx/faktura-pdf/config"
	"github.com/yzzyx/zapsentry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// IgnoreSync is a type that ignores calls to Sync()
// This can be used on special files like os.Stderr or os.Stdout that do not support fsync()-calls on Linux
type IgnoreSync struct {
	*os.File
}

// Sync is ignored for IgnoreSync
func (f IgnoreSync) Sync() error {
	return nil
}

// Convert config file strings to zap logging levels
func zapStringToLevel(s string) zapcore.LevelEnabler {
	var level zapcore.LevelEnabler
	switch strings.ToLower(s) {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "panic":
		level = zapcore.PanicLevel
	case "fatal":
		level = zapcore.FatalLevel
	default:
		level = zapcore.ErrorLevel
	}
	return level
}

func setupLogger(cfg config.Config) (*zap.Logger, error) {
	var coreList []zapcore.Core
	pe := zap.NewProductionEncoderConfig()
	fileEncoder := zapcore.NewJSONEncoder(pe)

	logFileName := cfg.Logging.Logfile

	if logFileName != "" {
		logFileName, err := strftime.Format(logFileName, time.Now().Local())
		if err != nil {
			return nil, err
		}

		logFile := zapcore.AddSync(&lumberjack.Logger{
			Filename: logFileName,
		})
		coreList = append(coreList, zapcore.NewCore(fileEncoder, logFile, zapStringToLevel(cfg.Logging.Level)))
	}

	pe.EncodeTime = zapcore.ISO8601TimeEncoder
	pe.StacktraceKey = "stacktrace"
	consoleEncoder := zapcore.NewConsoleEncoder(pe)
	coreList = append(coreList, zapcore.NewCore(consoleEncoder, IgnoreSync{os.Stderr}, zapStringToLevel(cfg.Logging.ConsoleLevel)))

	if cfg.Sentry.Enabled {
		client, err := sentry.NewClient(sentry.ClientOptions{
			Dsn: cfg.Sentry.DSN,
			// Skip default integrations, we set our own tags instead
			Integrations: func(defaultIntegrations []sentry.Integration) []sentry.Integration {
				integrations := defaultIntegrations[:0]
				for k := range defaultIntegrations {
					if defaultIntegrations[k].Name() == "Modules" ||
						defaultIntegrations[k].Name() == "Environment" ||
						defaultIntegrations[k].Name() == "ContextifyFrames" {
						continue
					}
					integrations = append(integrations, defaultIntegrations[k])
				}
				return integrations
			},
		})

		scope := sentry.NewScope()
		//scope.SetTag("version", VersionNumber)
		hostname, err := os.Hostname()
		if err != nil {
			return nil, err
		}

		scope.SetTag("server_name", hostname)
		scope.SetTag("device.arch", runtime.GOARCH)
		scope.SetTag("os.name", runtime.GOOS)

		scope.SetTag("runtime.name", "go")
		scope.SetTag("runtime.version", runtime.Version())
		hub := sentry.NewHub(client, scope)

		sentryCore, err := zapsentry.NewCore(hub, zapStringToLevel(cfg.Sentry.Level))
		if err != nil {
			return nil, err
		}
		coreList = append(coreList, sentryCore)
	}

	core := zapcore.NewTee(coreList...)
	return zap.New(core), nil
}
