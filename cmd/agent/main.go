package main

import (
	"github.com/pavelborisofff/go-metrics/internal/config"
	"github.com/pavelborisofff/go-metrics/internal/logger"
	"github.com/pavelborisofff/go-metrics/internal/storage"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log := logger.GetLogger()
	defer log.Sync()

	s := storage.NewAgentStorage()
	cfg, err := config.GetAgentConfig()
	if err != nil {
		log.Fatal("Error load config", zap.Error(err))
	}
	log.Debug("Agent Config", zap.Any("config", cfg))

	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)

	metricsCh := make(chan storage.AgentStorage)
	defer close(metricsCh)

	// Create a worker pool
	for i := 0; i < cfg.Agent.RateLimit; i++ {
		go func() {
			for {
				select {
				case <-exitSignal:
					return
				case <-metricsCh:
					err = s.SendMetrics(cfg.ServerAddr)
					if err != nil {
						log.Error("Error sending metrics", zap.Error(err))
					}

					err = s.BatchSend(cfg.ServerAddr)
					if err != nil {
						log.Error("Error sending batch metrics", zap.Error(err))
					}
				}
			}
		}()
	}

	// Collect metrics and send them to the channel
	go func() {
		pollTicker := time.NewTicker(time.Duration(cfg.Agent.PollInterval) * time.Second)
		reportTicker := time.NewTicker(time.Duration(cfg.Agent.ReportInterval) * time.Second)
		defer pollTicker.Stop()
		defer reportTicker.Stop()

		for {
			select {
			case <-exitSignal:
				return
			case <-pollTicker.C:
				s.UpdateMetrics()
			case <-reportTicker.C:
				metricsCh <- *s
			}
		}
	}()

	// New goroutine for collecting additional metrics
	go func() {
		for {
			select {
			case <-exitSignal:
				return
			default:
				v, _ := mem.VirtualMemory()
				c, _ := cpu.Percent(0, false)

				s.UpdateGauge("TotalMemory", storage.Gauge(v.Total))
				s.UpdateGauge("FreeMemory", storage.Gauge(v.Free))
				s.UpdateGauge("CPUutilization1", storage.Gauge(c[0]))

				time.Sleep(time.Duration(cfg.Agent.PollInterval) * time.Second)
			}
		}
	}()

	<-exitSignal
}
