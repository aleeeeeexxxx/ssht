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

	// Wrap core with broadcast hook
	broadcastCore := &broadcastCoreWrapper{Core: logger.Core()}
	Log = zap.New(broadcastCore).Sugar()
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

type broadcastCoreWrapper struct {
	zapcore.Core
}

func (c *broadcastCoreWrapper) With(fields []zapcore.Field) zapcore.Core {
	return &broadcastCoreWrapper{Core: c.Core.With(fields)}
}

func (c *broadcastCoreWrapper) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Core.Enabled(entry.Level) {
		return ce.AddCore(entry, c)
	}
	return ce
}

func (c *broadcastCoreWrapper) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	if broadcastFunc != nil {
		fieldMap := make(map[string]interface{})
		for _, f := range fields {
			switch f.Type {
			case zapcore.StringType:
				fieldMap[f.Key] = f.String
			case zapcore.Int64Type, zapcore.Int32Type, zapcore.Int16Type, zapcore.Int8Type:
				fieldMap[f.Key] = f.Integer
			case zapcore.Uint64Type, zapcore.Uint32Type, zapcore.Uint16Type, zapcore.Uint8Type:
				fieldMap[f.Key] = uint64(f.Integer)
			case zapcore.Float64Type:
				fieldMap[f.Key] = float64(f.Integer)
			case zapcore.Float32Type:
				fieldMap[f.Key] = float32(f.Integer)
			case zapcore.BoolType:
				fieldMap[f.Key] = f.Integer == 1
			case zapcore.ErrorType:
				if f.Interface != nil {
					fieldMap[f.Key] = f.Interface.(error).Error()
				}
			default:
				if f.Interface != nil {
					fieldMap[f.Key] = f.Interface
				}
			}
		}
		broadcastFunc(entry.Level.CapitalString(), entry.Message, fieldMap)
	}
	return c.Core.Write(entry, fields)
}
