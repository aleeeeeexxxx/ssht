package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.SugaredLogger
var broadcastFunc func(level, message string, fields map[string]interface{})

func Init(debug bool) {
	var config zap.Config

	if debug {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "time"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	logger, err := config.Build()
	if err != nil {
		os.Exit(1)
	}

	// Wrap with broadcast hook
	Log = logger.WithOptions(zap.Hooks(broadcastHook)).Sugar()
}

func Sync() {
	if Log != nil {
		Log.Sync()
	}
}

// SetBroadcastFunc sets the function to broadcast logs to WebSocket clients
func SetBroadcastFunc(fn func(level, message string, fields map[string]interface{})) {
	broadcastFunc = fn
}

func broadcastHook(entry zapcore.Entry) error {
	if broadcastFunc != nil {
		broadcastFunc(entry.Level.CapitalString(), entry.Message, nil)
	}
	return nil
}
