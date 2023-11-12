package logger

import (
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	instance = zap.NewNop()
	once     sync.Once
)

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

func GetLogger() *zap.Logger {
	once.Do(func() {
		logger, err := zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
		instance = logger
	})

	return instance
}

func LogHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		instance.Info("request",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.Int("status", ResponseData.status),
			zap.Int("size", ResponseData.size),
			zap.Duration("duration", duration),
			zap.String("Content-Type", r.Header.Get("Content-Type")),
			zap.String("Content-Encoding", r.Header.Get("Content-Encoding")),
			zap.String("Accept-Encoding", r.Header.Get("Accept-Encoding")),
		)
	})
}
