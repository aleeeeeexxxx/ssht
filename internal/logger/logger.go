package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.SugaredLogger
var broadcastFunc func(level, message string, fields map[string]interface{})
var atomicLevel zap.AtomicLevel

func Init(debug bool) {
	if debug {
		atomicLevel = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	} else {
		atomicLevel = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if debug {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		atomicLevel,
	)

	// Wrap core with broadcast hook
	broadcastCore := &broadcastCoreWrapper{Core: core}
	Log = zap.New(broadcastCore).Sugar()
}

func Sync() {
	if Log != nil {
		Log.Sync()
	}
}

// SetLevel changes the log level dynamically
func SetLevel(level string) error {
	var l zapcore.Level
	if err := l.UnmarshalText([]byte(level)); err != nil {
		return err
	}
	atomicLevel.SetLevel(l)
	return nil
}

// GetLevel returns the current log level
func GetLevel() string {
	return atomicLevel.Level().String()
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
