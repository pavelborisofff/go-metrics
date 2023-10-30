package storage

import (
	"encoding/json"
	"fmt"
	"github.com/pavelborisofff/go-metrics/internal/gzip"
	"log"
	"math/rand"
	"net/http"
	"runtime"
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

		s.SendMetric(CounterType, name, value, serverAddr)
	}

	for name, value := range s.GaugeStorage {
		m = Metrics{
			ID:    name,
			MType: GaugeType,
			Value: new(float64),
		}
		*m.Value = float64(value)

		s.SendJSONMetric(m, serverAddr)
	}

	return nil
}

func (s *AgentStorage) SendJSONMetric(m Metrics, serverAddr string) {
	data, err := json.Marshal(m)
	if err != nil {
		msg := fmt.Sprintf("Error: %s", err)
		log.Println(msg)
		return
	}

	compressedData, err := gzip.CompressData(data)
	if err != nil {
		msg := fmt.Sprintf("Error compressing metric data: %s", err)
		log.Println(msg)
		return
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/update/", serverAddr), compressedData)
	if err != nil {
		msg := fmt.Sprintf("Failed to send metric: %s", err)
		log.Println(msg)
		return
	}

	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Content-Encoding", "gzip")

	c := &http.Client{}
	res, err := c.Do(req)
	if err != nil {
		msg := fmt.Sprintf("Failed to send metric: %s", err)
		log.Println(msg)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("Failed to send metric: %s", res.Status)
		log.Println(msg)
		return
	}

	msg := fmt.Sprintf("Metric sent successfully: %s", string(data))
	log.Println(msg)
}

func (s *AgentStorage) SendMetric(metricType string, metricName string, metricValue interface{}, serverAddr string) {
	url := fmt.Sprintf("%s/update/%s/%s/%v", serverAddr, metricType, metricName, metricValue)

	res, err := http.Post(url, "text/plain", nil)
	if err != nil {
		msg := fmt.Sprintf("Failed to send metric: %s", err)
		log.Default().Println(msg)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("Failed to send metric: %s", res.Status)
		log.Default().Println(msg)
		return
	}

	msg := fmt.Sprintf("Metric sent successfully: %s", url)
	log.Default().Println(msg)
}

/*func (s *AgentStorage) SendJSONMetric(m Metrics, serverAddr string) {
	data, err := json.Marshal(m)
	if err != nil {
		msg := fmt.Sprintf("Error: %s", err)
		log.Println(msg)
		return
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/update/", serverAddr), bytes.NewBuffer(data))
	if err != nil {
		msg := fmt.Sprintf("Failed to send metric: %s", err)
		log.Println(msg)
		return
	}

	req.Header.Set("Content-Type", "text/plain")

	var res *http.Response
	err = retry.Do(
		func() error {
			c := &http.Client{}
			res, err = c.Do(req)
			return err
		},
		retry.Attempts(5),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("Retry #%d: %s", n, err)
		}),
	)
	if err != nil {
		msg := fmt.Sprintf("Failed to send metric: %s", err)
		log.Println(msg)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("Failed to send metric: %s", res.Status)
		log.Println(msg)
		return
	}

	msg := fmt.Sprintf("Metric sent successfully: %s", string(data))
	log.Println(msg)
}

func (s *AgentStorage) SendMetric(metricType string, metricName string, metricValue interface{}, serverAddr string) {
	url := fmt.Sprintf("%s/update/%s/%s/%v", serverAddr, metricType, metricName, metricValue)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		msg := fmt.Sprintf("Failed to send metric: %s", err)
		log.Println(msg)
		return
	}

	req.Header.Set("Content-Type", "text/plain")

	var res *http.Response
	err = retry.Do(
		func() error {
			c := &http.Client{}
			res, err = c.Do(req)
			return err
		},
		retry.Attempts(5),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("Retry #%d: %s", n, err)
		}),
	)
	if err != nil {
		msg := fmt.Sprintf("Failed to send metric: %s", err)
		log.Println(msg)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("Failed to send metric: %s", res.Status)
		log.Default().Println(msg)
		return
	}

	msg := fmt.Sprintf("Metric sent successfully: %s", url)
	log.Default().Println(msg)
}*/
