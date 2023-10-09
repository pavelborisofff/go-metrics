package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
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

	defer res.Body.Close()
}

func ParseFlags() {
	var (
		serverAddrFlag     string
		pollIntervalFlag   int
		reportIntervalFlag int
	)

	flag.StringVar(&serverAddrFlag, "a", serverAddrDef, "Server address")
	flag.IntVar(&pollIntervalFlag, "p", pollIntervalDef, "Poll interval")
	flag.IntVar(&reportIntervalFlag, "r", reportIntervalDef, "Report interval")
	flag.Parse()

	serverAddr = fmt.Sprintf(`http://%s/update`, serverAddrFlag)
	pollInterval = pollIntervalFlag
	reportInterval = reportIntervalFlag
}

func main() {
	storage := NewMemStorage()
	ParseFlags()

	go func() {
		for {
			storage.UpdateMetrics()
			time.Sleep(time.Duration(pollInterval) * time.Second)
		}
	}()

	for {
		err := storage.SendMetrics()
		if err != nil {
			return
		}
		time.Sleep(time.Duration(reportInterval) * time.Second)
	}
}
