package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"time"
)

const (
	pollIntervalDef   = 2
	reportIntervalDef = 10
	serverAddrDef     = `localhost:8080`
)

type Gauge float64
type Counter uint64

var (
	pollInterval   int
	reportInterval int
	serverAddr     string
	memStatsNames  = []string{
		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
	}
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

	k := reflect.TypeOf(m)
	v := reflect.ValueOf(m)

	for _, name := range memStatsNames {
		field, ok := k.FieldByName(name)

		if !ok {
			continue
		}

		value := v.FieldByName(name)

		switch value.Kind() {
		case reflect.Uint64:
			s.UpdateGauge(field.Name, Gauge(value.Uint()))
		case reflect.Uint32:
			s.UpdateGauge(field.Name, Gauge(value.Uint()))
		case reflect.Float64:
			s.UpdateGauge(field.Name, Gauge(value.Float()))
		default:
			log.Default().Printf("Unknown type: %s", value.Kind())
			continue
		}
	}

	s.IncrementCounter(`PollCount`, 1)
	s.IncrementCounter(`RandomValue`, Counter(rand.Uint64()))
}

func (s *MemStorage) IncrementCounter(name string, value Counter) {
	s.CounterStorage[name] += value
}

func (s *MemStorage) SendMetrics() error {
	for name, value := range s.CounterStorage {
		s.SendMetric(`counter`, name, value)
	}

	for name, value := range s.GaugeStorage {
		s.SendMetric(`gauge`, name, value)
	}

	return nil
}

func (s *MemStorage) SendMetric(metricType string, metricName string, metricValue interface{}) {
	url := fmt.Sprintf(`%s/%s/%s/%v`, serverAddr, metricType, metricName, metricValue)

	res, err := http.Post(url, `text/plain`, nil)

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
		pollIntervalFlag   int
		reportIntervalFlag int
	)

	flag.StringVar(&serverAddrFlag, "a", serverAddrDef, "Server address")
	flag.IntVar(&pollIntervalFlag, "p", pollIntervalDef, "Poll interval")
	flag.IntVar(&reportIntervalFlag, "r", reportIntervalDef, "Report interval")
	flag.Parse()

	//serverAddrEnv, exists := os.LookupEnv("ADDRESS")
	serverAddrEnv, exists := os.LookupEnv("ADDRESS")
	if exists {
		serverAddrFlag = serverAddrEnv
	}
	serverAddr = fmt.Sprintf(`http://%s/update`, serverAddrFlag)

	pollIntervalEnv, exists := os.LookupEnv("POLL_INTERVAL")
	if exists {
		pollIntervalFlag, err = strconv.Atoi(pollIntervalEnv)
		if err != nil {
			log.Fatal(err)
		}
	}
	pollInterval = pollIntervalFlag

	if pollInterval < 1 {
		log.Fatal("Poll interval must be greater than 0")
	}

	reportIntervalEnv, exists := os.LookupEnv("REPORT_INTERVAL")
	if exists {
		reportIntervalFlag, err = strconv.Atoi(reportIntervalEnv)
		if err != nil {
			log.Fatal(err)
		}
	}
	reportInterval = reportIntervalFlag

	if reportInterval < 1 {
		log.Fatal("Report interval must be greater than 0")
	}

	msg := fmt.Sprintf("\nServer address: %s\nPoll interval: %v\nReport interval: %v", serverAddr, pollInterval, reportInterval)
	log.Default().Println(msg)
}

func main() {
	storage := NewMemStorage()
	ParseFlags()

	pollTicker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer pollTicker.Stop()
	reportTicker := time.NewTicker(time.Duration(reportInterval) * time.Second)
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
