package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var Log = zap.NewNop()

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func InitLogger() error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("cannot initialize zap")
	}
	defer logger.Sync()

	Log = logger
	return nil
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = InitLogger()
		start := time.Now()

		ResponseData := &responseData{
			status: 0,
			size:   0,
		}

		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   ResponseData,
		}
		next.ServeHTTP(&lw, r)

		duration := time.Since(start)

		Log.Info("request",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.Int("status", ResponseData.status),
			zap.Int("size", ResponseData.size),
			zap.Duration("duration", duration),
		)
	})
}
