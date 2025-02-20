package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}

	// implement http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // embed original http.ResponseWriter
		responseData        *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// write response using the original http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // catch size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// write status code using the original http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // catch status code
}

// Log is a Singleton.
// Only the Initialize function can modify Log.
var Log *zap.Logger = zap.NewNop()

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

func WithLogging(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rd := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // embed original http.ResponseWriter
			responseData:   rd,
		}

		h(&lw, r)

		duration := time.Since(start)

		Log.Info("got incoming HTTP request",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Int("status", rd.status),
			zap.Int("size", rd.size),
			zap.Duration("duration", duration),
		)
	}
}
