package logger

import (
	"go.uber.org/zap"
)

// Log is a Singleton.
// Only the Initialize function can modify Log.
var Log = zap.NewNop()

// Initialize sets up the singleton Log with the specified log level.
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zl
	return nil
}
