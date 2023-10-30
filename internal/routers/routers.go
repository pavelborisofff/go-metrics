package routers

import (
	"github.com/go-chi/chi/v5"

	"github.com/pavelborisofff/go-metrics/internal/handlers"
	"github.com/pavelborisofff/go-metrics/internal/logger"
)

func InitRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(logger.Middleware)

	r.Get("/", handlers.MainHandler)
	r.Post("/update/{metric-type}/{metric-name}/{metric-value}", handlers.UpdateHandler)
	r.Get("/value/{metric-type}/{metric-name}", handlers.ValueHandler)
	r.Post("/update/", handlers.UpdateJSONHandler)
	r.Post("/value/", handlers.ValueJSONHandler)
	r.Get("/metrics", handlers.MetricsHandler)

	return r
}
