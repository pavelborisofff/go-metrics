package storage

import (
	"encoding/json"
	"fmt"
	"github.com/pavelborisofff/go-metrics/internal/config"
	"github.com/pavelborisofff/go-metrics/internal/gzip"
	"github.com/pavelborisofff/go-metrics/internal/hash"
	"github.com/pavelborisofff/go-metrics/internal/logger"
	"github.com/pavelborisofff/go-metrics/internal/retrying"
	"go.uber.org/zap"
	"math/rand"
	"net/http"
	"runtime"
)

var (
	log    = logger.GetLogger()
	client = &http.Client{}
	cfg, _ = config.GetAgentConfig()
)

type AgentStorage struct {
	MemStorage
}

func NewAgentStorage() *AgentStorage {
	return &AgentStorage{
		MemStorage: *NewMemStorage(),
	}
}

func (s *AgentStorage) UpdateMetrics() {
	var m runtime.MemStats

	runtime.ReadMemStats(&m)
	s.UpdateGauge("Alloc", Gauge(m.Alloc))
	s.UpdateGauge("BuckHashSys", Gauge(m.BuckHashSys))
	s.UpdateGauge("Frees", Gauge(m.Frees))
	s.UpdateGauge("GCCPUFraction", Gauge(m.GCCPUFraction))
	s.UpdateGauge("GCSys", Gauge(m.GCSys))
	s.UpdateGauge("HeapAlloc", Gauge(m.HeapAlloc))
	s.UpdateGauge("HeapIdle", Gauge(m.HeapIdle))
	s.UpdateGauge("HeapInuse", Gauge(m.HeapInuse))
	s.UpdateGauge("HeapObjects", Gauge(m.HeapObjects))
	s.UpdateGauge("HeapReleased", Gauge(m.HeapReleased))
	s.UpdateGauge("HeapSys", Gauge(m.HeapSys))
	s.UpdateGauge("LastGC", Gauge(m.LastGC))
	s.UpdateGauge("Lookups", Gauge(m.Lookups))
	s.UpdateGauge("MCacheInuse", Gauge(m.MCacheInuse))
	s.UpdateGauge("MCacheSys", Gauge(m.MCacheSys))
	s.UpdateGauge("MSpanInuse", Gauge(m.MSpanInuse))
	s.UpdateGauge("MSpanSys", Gauge(m.MSpanSys))
	s.UpdateGauge("Mallocs", Gauge(m.Mallocs))
	s.UpdateGauge("NextGC", Gauge(m.NextGC))
	s.UpdateGauge("NumForcedGC", Gauge(m.NumForcedGC))
	s.UpdateGauge("NumGC", Gauge(m.NumGC))
	s.UpdateGauge("OtherSys", Gauge(m.OtherSys))
	s.UpdateGauge("PauseTotalNs", Gauge(m.PauseTotalNs))
	s.UpdateGauge("StackInuse", Gauge(m.StackInuse))
	s.UpdateGauge("StackSys", Gauge(m.StackSys))
	s.UpdateGauge("Sys", Gauge(m.Sys))
	s.UpdateGauge("TotalAlloc", Gauge(m.TotalAlloc))
	s.UpdateGauge("RandomValue", Gauge(rand.Uint64()))

	s.IncrementCounter("PollCount", 1)

	// gopsutil/v3 metrics
	//v, _ := mem.VirtualMemory()
	//c, _ := cpu.Percent(0, true)
	//
	//s.UpdateGauge("TotalMemory", Gauge(v.Total))
	//s.UpdateGauge("FreeMemory", Gauge(v.Free))
	//s.UpdateGauge("CPUutilization1", Gauge(c[0]))
}

func (s *AgentStorage) SendMetrics(serverAddr string) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	for name, value := range s.CounterStorage {
		s.SendMetric(CounterType, name, value, serverAddr)
	}

	for name, value := range s.GaugeStorage {
		s.SendMetric(GaugeType, name, value, serverAddr)
	}

	return nil
}

func (s *AgentStorage) SendJSONMetrics(serverAddr string) error {
	var m Metrics

	for name, value := range s.CounterStorage {
		m = Metrics{
			ID:    name,
			MType: CounterType,
			Delta: new(int64),
		}
		*m.Delta = int64(value)

		err := s.SendJSONMetric(m, serverAddr)
		if err != nil {
			return err
		}
	}

	for name, value := range s.GaugeStorage {
		m = Metrics{
			ID:    name,
			MType: GaugeType,
			Value: new(float64),
		}
		*m.Value = float64(value)

		err := s.SendJSONMetric(m, serverAddr)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *AgentStorage) batchSendMetrics(m []Metrics, serverAddr string) error {
	data, err := json.Marshal(m)
	if err != nil {
		log.Error("Error marshaling JSON batch data", zap.Error(err))
		return err
	}

	compressedData, err := gzip.CompressData(data)
	if err != nil {
		log.Error("Error compressing JSON batch data", zap.Error(err))
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/updates/", serverAddr), compressedData)
	if err != nil {
		log.Error("Error creating batch request JSON", zap.Error(err))
		return err
	}

	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Content-Encoding", "gzip")

	if cfg.UseHashKey {
		hashed := hash.Make(data, cfg.HashKey)
		log.Info("Hashed", zap.String("hash", hashed))
		req.Header.Set("HashSHA256", hashed)
	}

	resp, err := retrying.Request(client, req)
	if err != nil {
		log.Error("Error sending batch request JSON", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error("Error sending batch request JSON", zap.String("status", resp.Status))
		return err
	}

	log.Info("Batch JSON sent successfully")
	return nil
}

func (s *AgentStorage) BatchSend(serverAddr string) error {
	var m []Metrics
	s.Mu.Lock()
	defer s.Mu.Unlock()

	for k, v := range s.CounterStorage {
		delta := int64(v)
		m = append(m, Metrics{
			ID:    k,
			MType: CounterType,
			Delta: &delta,
		})
	}

	for k, v := range s.GaugeStorage {
		value := float64(v)
		m = append(m, Metrics{
			ID:    k,
			MType: GaugeType,
			Value: &value,
		})
	}

	err := s.batchSendMetrics(m, serverAddr)
	if err != nil {
		return err
	}

	return nil
}

func (s *AgentStorage) SendJSONMetric(m Metrics, serverAddr string) error {
	data, err := json.Marshal(m)
	if err != nil {
		log.Error("Error marshaling JSON data", zap.Error(err))
		return err
	}

	compressedData, err := gzip.CompressData(data)
	if err != nil {
		log.Error("Error compressing JSON data", zap.Error(err))
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/update/", serverAddr), compressedData)
	if err != nil {
		log.Error("Error creating request JSON", zap.Error(err))
		return err
	}

	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Content-Encoding", "gzip")

	if cfg.UseHashKey {
		hashed := hash.Make(data, cfg.HashKey)
		log.Info("Hashed", zap.String("hash", hashed))
		req.Header.Set("HashSHA256", hashed)
	}

	res, err := retrying.Request(client, req)
	if err != nil {
		log.Error("Failed to send metric", zap.Error(err))
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Error("Error sending JSON", zap.String("status", res.Status))
		return err
	}

	log.Info("JSON sent successfully", zap.ByteString("data", data))
	return nil
}

func (s *AgentStorage) SendMetric(metricType string, metricName string, metricValue interface{}, serverAddr string) error {
	url := fmt.Sprintf("%s/update/%s/%s/%v", serverAddr, metricType, metricName, metricValue)

	res, err := http.Post(url, "text/plain", nil)
	if err != nil {
		log.Debug("Failed to send metric", zap.Error(err))
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Debug("Error sending metric", zap.String("status", res.Status))
		return err
	}

	log.Info("Metric sent successfully", zap.String("url", url))
	return nil
}

//func (s *AgentStorage) UpdateMetrics(cfg *config.Config, metricsCh chan AgentStorage) error {
//	pollTicker := time.NewTicker(time.Duration(cfg.Agent.PollInterval) * time.Second)
//	reportTicker := time.NewTicker(time.Duration(cfg.Agent.ReportInterval) * time.Second)
//
//	for {
//		select {
//		case <-pollTicker.C:
//			s.updateMetrics()
//		case <-reportTicker.C:
//			metricsCh <- *s
//			metricsCh <- *s
//		}
//	}
//}

//func CollectMetrics(cfg *config.Config, metricsCh chan MemStorage) {
//	s := NewAgentStorage()
//	pollTicker := time.NewTicker(time.Duration(cfg.Agent.PollInterval) * time.Second)
//	//defer pollTicker.Stop()
//
//	for {
//		select {
//		case <-pollTicker.C:
//			s.UpdateMetrics()
//			metricsCh <- s.MemStorage
//		}
//	}
//}
