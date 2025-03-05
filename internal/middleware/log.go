package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
	"shorturl/internal/logger"
)

type (
	responseData struct {
		status int
		size   int
	}

	// LoggingResponseWriter implements http.ResponseWriter
	LoggingResponseWriter struct {
		http.ResponseWriter // embed original http.ResponseWriter
		ResponseData        *responseData
	}
)

func (r *LoggingResponseWriter) Write(b []byte) (int, error) {
	// write response using the original http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.ResponseData.size += size // catch size
	return size, err
}

func (r *LoggingResponseWriter) WriteHeader(statusCode int) {
	// write status code using the original http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.ResponseData.status = statusCode // catch status code
}

func WithLogging(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rd := &responseData{
			status: 0,
			size:   0,
		}
		lw := LoggingResponseWriter{
			ResponseWriter: w, // embed original http.ResponseWriter
			ResponseData:   rd,
		}

		h(&lw, r)

		duration := time.Since(start)

		logger.Log.Info("got incoming HTTP request",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Int("status", rd.status),
			zap.Int("size", rd.size),
			zap.Duration("duration", duration),
		)
	}
}
