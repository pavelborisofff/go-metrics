package routers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/pavelborisofff/go-metrics/internal/handlers"
	"github.com/pavelborisofff/go-metrics/internal/logger"
	"github.com/pavelborisofff/go-metrics/internal/storage"
)

func InitRouter(s *storage.MemStorage) *chi.Mux {
	r := chi.NewRouter()
	r.Use(logger.Middleware)

	r.Get("/", func(res http.ResponseWriter, req *http.Request) {
		handlers.MainHandler(res, req, s)
	})
	r.Post("/update/{metric-type}/{metric-name}/{metric-value}", func(res http.ResponseWriter, req *http.Request) {
		handlers.UpdateHandler(res, req, s)
	})
	r.Get("/value/{metric-type}/{metric-name}", func(res http.ResponseWriter, req *http.Request) {
		handlers.ValueHandler(res, req, s)
	})
	r.Get("/metrics", func(res http.ResponseWriter, req *http.Request) {
		handlers.MetricsHandler(res, req, s)
	})

	return r
}
