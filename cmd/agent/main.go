package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"time"
)

const (
	pollIntervalDef   = time.Duration(2)
	reportIntervalDef = time.Duration(10)
	serverAddrDef     = "localhost:8080"
)

type Gauge float64
type Counter uint64

var (
	pollInterval   time.Duration
	reportInterval time.Duration
	serverAddr     string
)

type MemStorage struct {
	CounterStorage map[string]Counter `json:"counter"`
	GaugeStorage   map[string]Gauge   `json:"gauge"`
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		CounterStorage: make(map[string]Counter),
		GaugeStorage:   make(map[string]Gauge),
	}
}

func (s *MemStorage) UpdateGauge(name string, value Gauge) {
	s.GaugeStorage[name] = value
}

func (s *MemStorage) UpdateMetrics() {
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

	s.IncrementCounter("PollCount", 1)
	s.IncrementCounter("RandomValue", Counter(rand.Uint64()))
}

func (s *MemStorage) IncrementCounter(name string, value Counter) {
	s.CounterStorage[name] += value
}

func (s *MemStorage) SendMetrics() error {
	for name, value := range s.CounterStorage {
		s.SendMetric("counter", name, value)
	}

	for name, value := range s.GaugeStorage {
		s.SendMetric("gauge", name, value)
	}

	return nil
}

func (s *MemStorage) SendMetric(metricType string, metricName string, metricValue interface{}) {
	url := fmt.Sprintf("%s/%s/%s/%v", serverAddr, metricType, metricName, metricValue)

	res, err := http.Post(url, "text/plain", nil)

	if err != nil {
		msg := fmt.Sprintf("Failed to send metric: %s", err)
		log.Default().Println(msg)
		return
	}

	if res.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("Failed to send metric: %s", res.Status)
		log.Default().Println(msg)
		return
	}

	msg := fmt.Sprintf("Metric sent successfully: %s", url)
	log.Default().Println(msg)

	res.Body.Close()
}

func ParseFlags() {
	var (
		err                error
		serverAddrFlag     string
		pollIntervalFlag   time.Duration
		reportIntervalFlag time.Duration
	)

	msg2 := fmt.Sprintf("\n\nPoll interval: %v\nReport interval: %v", pollIntervalFlag, pollInterval)
	log.Default().Println(msg2)

	flag.StringVar(&serverAddrFlag, "a", serverAddrDef, "Server address")
	flag.DurationVar(&pollIntervalFlag, "p", pollIntervalDef*time.Second, "Poll interval")
	flag.DurationVar(&reportIntervalFlag, "r", reportIntervalDef*time.Second, "Report interval")

	flag.Parse()

	//serverAddrEnv, exists := os.LookupEnv("ADDRESS")
	serverAddrEnv, exists := os.LookupEnv("ADDRESS")
	if exists {
		serverAddrFlag = serverAddrEnv
	}
	serverAddr = fmt.Sprintf("http://%s/update", serverAddrFlag)

	pollIntervalEnv, exists := os.LookupEnv("POLL_INTERVAL")
	if exists {
		pollIntervalFlag, err = time.ParseDuration(pollIntervalEnv + "s")
		if err != nil {
			log.Fatal(err)
		}
	}
	pollInterval = pollIntervalFlag

	if pollInterval < 1*time.Second {
		log.Fatal("Poll interval must be >= 1s")
	}

	reportIntervalEnv, exists := os.LookupEnv("REPORT_INTERVAL")
	if exists {
		reportIntervalFlag, err = time.ParseDuration(reportIntervalEnv + "s")
		if err != nil {
			log.Fatal(err)
		}
	}
	reportInterval = reportIntervalFlag

	if reportInterval < 1*time.Second {
		log.Fatal("Report interval must be >= 1s")
	}

	msg := fmt.Sprintf("\nServer address: %s\nPoll interval: %v\nReport interval: %v", serverAddr, pollInterval, reportInterval)
	log.Default().Println(msg)
}

func main() {
	storage := NewMemStorage()
	ParseFlags()

	pollTicker := time.NewTicker(pollInterval)
	defer pollTicker.Stop()
	reportTicker := time.NewTicker(reportInterval)
	defer reportTicker.Stop()

	for {
		select {
		case <-pollTicker.C:
			storage.UpdateMetrics()
		case <-reportTicker.C:
			err := storage.SendMetrics()
			if err != nil {
				return
			}
		}
	}
}
