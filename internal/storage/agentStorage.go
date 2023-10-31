package storage

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"math/rand"
	"net/http"
	"runtime"

	"github.com/pavelborisofff/go-metrics/internal/gzip"
	"github.com/pavelborisofff/go-metrics/internal/logger"
)

var (
	log = logger.Log
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
}

func (s *AgentStorage) SendMetrics(serverAddr string) error {
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

func (s *AgentStorage) SendJSONMetric(m Metrics, serverAddr string) error {
	data, err := json.Marshal(m)
	if err != nil {
		log.Error("Error marshaling metric data", zap.Error(err))
		return err
	}

	compressedData, err := gzip.CompressData(data)
	if err != nil {
		log.Error("Error compressing metric data", zap.Error(err))
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/update/", serverAddr), compressedData)
	if err != nil {
		log.Error("Error creating request", zap.Error(err))
		return err
	}

	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Content-Encoding", "gzip")

	c := &http.Client{}
	res, err := c.Do(req)
	if err != nil {
		log.Error("Failed to send metric", zap.Error(err))
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Error("Error sending metric", zap.String("status", res.Status))
		return err
	}

	log.Debug("Metric sent successfully", zap.ByteString("data", data))
	return nil
}

func (s *AgentStorage) SendMetric(metricType string, metricName string, metricValue interface{}, serverAddr string) {
	url := fmt.Sprintf("%s/update/%s/%s/%v", serverAddr, metricType, metricName, metricValue)

	res, err := http.Post(url, "text/plain", nil)
	if err != nil {
		log.Debug("Failed to send metric", zap.Error(err))
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Debug("Error sending metric", zap.String("status", res.Status))
		return
	}

	log.Debug("Metric sent successfully", zap.String("url", url))
}
