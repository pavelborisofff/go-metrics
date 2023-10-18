package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"

	"github.com/pavelborisofff/go-metrics/internal/logger"
)

type Gauge float64
type Counter uint64

const serverAddrDef = "localhost:8080"

//go:embed templates/metrics.html
var htmlMetrics string

var ServerAddr string

type MemStorage struct {
	CounterStorage map[string]Counter `json:"counter"`
	GaugeStorage   map[string]Gauge   `json:"gauge"`
	mu             *sync.Mutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		CounterStorage: make(map[string]Counter),
		GaugeStorage:   make(map[string]Gauge),
		mu:             &sync.Mutex{},
	}
}

func (s *MemStorage) UpdateGauge(name string, value Gauge) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.GaugeStorage[name] = value
}

func (s *MemStorage) IncrementCounter(name string, value Counter) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.CounterStorage[name] += value
}

func mainHandler(res http.ResponseWriter, req *http.Request, storage *MemStorage) {
	tmpl := template.Must(template.New("metrics").Parse(htmlMetrics))
	err := tmpl.Execute(res, storage)

	if err != nil {
		log.Println(err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

}

func updateHandler(res http.ResponseWriter, req *http.Request, storage *MemStorage) {
	metricType := chi.URLParam(req, "metric-type")
	metricName := chi.URLParam(req, "metric-name")
	metricValue := chi.URLParam(req, "metric-value")

	switch metricType {
	case `counter`:
		v, err := strconv.ParseUint(metricValue, 10, 64)

		if err != nil {
			msg := fmt.Sprintf("Bad Counter's value: %s", metricValue)
			log.Println(msg)
			http.Error(res, msg, http.StatusBadRequest)
			return
		}

		storage.IncrementCounter(metricName, Counter(v))
		msg := fmt.Sprintf("Counter %s shanged to %d", metricName, v)
		log.Println(msg)

	case `gauge`:
		v, err := strconv.ParseFloat(metricValue, 64)

		if err != nil {
			msg := fmt.Sprintf("Bad metric's value: %s %s", metricName, metricValue)
			log.Println(msg)
			http.Error(res, msg, http.StatusBadRequest)
			return
		}

		storage.UpdateGauge(metricName, Gauge(v))
		msg := fmt.Sprintf("Gauge %s updated to %f", metricName, v)
		log.Println(msg)

	default:
		msg := fmt.Sprintf("Bad metric's type: %s", metricType)
		log.Println(msg)
		http.Error(res, msg, http.StatusBadRequest)
	}

	res.WriteHeader(http.StatusOK)
}

func metricsHandler(res http.ResponseWriter, req *http.Request, storage *MemStorage) {
	data, err := json.Marshal(storage)
	if err != nil {
		msg := fmt.Sprintf("Error: %s", err)
		log.Println(msg)
		http.Error(res, msg, http.StatusInternalServerError)
		return
	}
	log.Default().Println(string(data))
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	write, err := res.Write(data)

	if err != nil {
		msg := fmt.Sprintf("Error: %s", err)
		http.Error(res, msg, http.StatusInternalServerError)
		log.Fatal(msg)
		return
	}

	log.Println(write)
}

func valueHandler(res http.ResponseWriter, req *http.Request, storage *MemStorage) {
	metricType := chi.URLParam(req, "metric-type")
	metricName := chi.URLParam(req, "metric-name")

	switch metricType {
	case "counter":
		if v, ok := storage.CounterStorage[metricName]; ok {
			io.WriteString(res, fmt.Sprintf("%v", v))
		} else {
			http.Error(res, "Not found", http.StatusNotFound)
		}
	case "gauge":
		if v, ok := storage.GaugeStorage[metricName]; ok {
			io.WriteString(res, fmt.Sprintf("%v", v))
		} else {
			http.Error(res, "Not found", http.StatusNotFound)
		}
	default:
		http.Error(res, "Bad metric's type", http.StatusBadRequest)
	}
}

func ParseFlags() {
	var (
		serverAddrFlag string
	)
	flag.StringVar(&serverAddrFlag, "a", serverAddrDef, "Server address")
	flag.Parse()

	serverAddrEnv, exists := os.LookupEnv("ADDRESS")
	if exists {
		serverAddrFlag = serverAddrEnv
	}

	ServerAddr = serverAddrFlag
	msg := fmt.Sprintf("\nServer address: %s", serverAddrFlag)
	log.Println(msg)
}

func main() {
	ParseFlags()
	storage := NewMemStorage()

	r := chi.NewRouter()
	r.Use(logger.Middleware)

	r.Get("/", func(res http.ResponseWriter, req *http.Request) {
		mainHandler(res, req, storage)
	})
	r.Post("/update/{metric-type}/{metric-name}/{metric-value}", func(res http.ResponseWriter, req *http.Request) {
		updateHandler(res, req, storage)
	})
	r.Get("/value/{metric-type}/{metric-name}", func(res http.ResponseWriter, req *http.Request) {
		valueHandler(res, req, storage)
	})
	r.Get("/metrics", func(res http.ResponseWriter, req *http.Request) {
		metricsHandler(res, req, storage)
	})

	log.Fatal(http.ListenAndServe(ServerAddr, r))
}
