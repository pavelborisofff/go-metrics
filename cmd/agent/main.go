package main

import (
	"go.uber.org/zap"
	"time"

	"github.com/pavelborisofff/go-metrics/internal/config"
	"github.com/pavelborisofff/go-metrics/internal/logger"
	"github.com/pavelborisofff/go-metrics/internal/storage"
)

func main() {
	log := logger.GetLogger()
	defer log.Sync()

	s := storage.NewAgentStorage()
	cfg, err := config.GetAgentConfig()
	if err != nil {
		log.Fatal("Error load config", zap.Error(err))
	}
	log.Debug("Config", zap.Any("config", cfg))

	pollTicker := time.NewTicker(time.Duration(cfg.Agent.PollInterval) * time.Second)
	defer pollTicker.Stop()
	reportTicker := time.NewTicker(time.Duration(cfg.Agent.ReportInterval) * time.Second)
	defer reportTicker.Stop()

	for {
		select {
		case <-pollTicker.C:
			s.UpdateMetrics()
		case <-reportTicker.C:
			err = s.SendJSONMetrics(cfg.ServerAddr)
			if err != nil {
				log.Error("Error sending metrics", zap.Error(err))
			}

			err = s.BatchSend(cfg.ServerAddr)
			if err != nil {
				log.Error("Error sending batch metrics", zap.Error(err))
			}
		}
	}
}
