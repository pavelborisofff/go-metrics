package main

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/pavelborisofff/go-metrics/internal/config"
	"github.com/pavelborisofff/go-metrics/internal/db"
	"github.com/pavelborisofff/go-metrics/internal/logger"
	"github.com/pavelborisofff/go-metrics/internal/routers"
	"github.com/pavelborisofff/go-metrics/internal/storage"
)

func main() {
	log := logger.GetLogger()
	defer log.Sync()

	s := storage.NewMemStorage()
	cfg, err := config.GetServerConfig()
	if err != nil {
		log.Fatal("Error load config", zap.Error(err))
	}

	switch cfg.Server.DBConn {
	case "":
		if cfg.Server.Restore {
			if err = s.FromFile(cfg.Server.FileStore); err != nil {
				log.Fatal("Error restore metrics", zap.Error(err))
			}
			log.Info("Metrics restored from file")
		}
	default:
		if err = db.InitDB(cfg.Server.DBConn); err != nil {
			log.Fatal("Error init DB", zap.Error(err))
			cfg.Server.DBConn = ""
		}
		defer db.DB.Close(context.Background())

		if err = db.CreateTables(); err != nil {
			log.Error("Error create tables", zap.Error(err))
		}

		log.Debug("DB init")

		if cfg.Server.Restore {
			if err = db.FromDatabase(s); err != nil {
				log.Fatal("Error restore metrics", zap.Error(err))
			}
			log.Info("Metrics restored from DB")
		}
	}

	go func() {
		if cfg.Server.SaveInterval <= 0 {
			return
		}

		ticker := time.NewTicker(time.Duration(cfg.Server.SaveInterval) * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			switch cfg.Server.DBConn {
			case "":
				if cfg.Server.FileStore == "" {
					return
				}
				if err := s.ToFile(cfg.Server.FileStore); err != nil {
					log.Fatal("Error saving metrics", zap.Error(err))
				}
				log.Debug("Metrics saved to file")
			default:
				if err := db.ToDatabase(s); err != nil {
					log.Error("Error saving metrics to Database", zap.Error(err))
				}
				log.Debug("Metrics saved to DB")
			}
		}
	}()

	r := routers.InitRouter()
	log.Fatal("Server error", zap.Error(http.ListenAndServe(cfg.ServerAddr, r)))
}
