package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type gauge float64
type counter int64

type MemStorage struct {
	CounterStorage map[string]counter `json:"counter"`
	GaugeStorage   map[string]gauge   `json:"gauge"`
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		CounterStorage: make(map[string]counter),
		GaugeStorage:   make(map[string]gauge),
	}
}

func (s *MemStorage) UpdateGauge(name string, value gauge) {
	s.GaugeStorage[name] = value
}

func (s *MemStorage) IncrementCounter(name string, value counter) {
	s.CounterStorage[name] += value
}

func middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		msg := fmt.Sprintf("From: %s, Path: %s, Method: %s", req.RemoteAddr, req.URL.Path, req.Method)
		log.Println(msg)
		next.ServeHTTP(res, req)
	}
}

func mainHandler(res http.ResponseWriter, _ *http.Request) {
	log.Println(`Bad request`)
	res.WriteHeader(http.StatusBadRequest)
}

func updateHandler(res http.ResponseWriter, req *http.Request, storage *MemStorage) {
	if req.Method != http.MethodPost {
		msg := fmt.Sprintf("Method not allowed: %s", req.Method)
		log.Println(msg)
		http.Error(res, msg, http.StatusMethodNotAllowed)
	}

	// server-address/update/metrics-type/metrics-name/metrics-value
	parts := strings.Split(req.URL.Path, `/`)

	if len(parts) != 5 {
		msg := fmt.Sprintf("Not found: %s", req.URL.Path)
		log.Println(msg)
		res.WriteHeader(http.StatusNotFound)
		return
	}

	metricType := parts[2]
	metricName := parts[3]
	metricValue := parts[4]

	switch metricType {
	case `counter`:
		v, err := strconv.ParseInt(metricValue, 10, 64)

		if err != nil {
			msg := fmt.Sprintf("Bad counter's value: %s", metricValue)
			log.Println(msg)
			http.Error(res, msg, http.StatusBadRequest)
			return
		}

		storage.IncrementCounter(metricName, counter(v))
		msg := fmt.Sprintf("Counter %s incremented by %d", metricName, v)
		log.Println(msg)

	case `gauge`:
		v, err := strconv.ParseFloat(metricValue, 64)

		if err != nil {
			msg := fmt.Sprintf("Bad metric's value: %s %s", metricName, metricValue)
			log.Println(msg)
			http.Error(res, msg, http.StatusBadRequest)
			return
		}

		storage.UpdateGauge(metricName, gauge(v))
		msg := fmt.Sprintf("Gauge %s updated to %f", metricName, v)
		log.Println(msg)

	default:
		msg := fmt.Sprintf("Bad metric's name: %s", metricType)
		log.Println(msg)
		http.Error(res, msg, http.StatusBadRequest)
	}

	res.WriteHeader(http.StatusOK)
}

func metricsHandler(res http.ResponseWriter, req *http.Request, storage *MemStorage) {
	if req.Method != http.MethodGet {
		msg := fmt.Sprintf("Method not allowed: %s", req.Method)
		log.Println(msg)
		http.Error(res, `Method not allowed`, http.StatusMethodNotAllowed)
	}

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
		return
	}
	log.Println(write)
}

func main() {
	storage := NewMemStorage()

	mux := http.NewServeMux()
	mux.HandleFunc(`/`, middleware(mainHandler))
	mux.HandleFunc(`/update/`, middleware(func(res http.ResponseWriter, req *http.Request) {
		updateHandler(res, req, storage)
	}))
	mux.HandleFunc(`/metrics`, middleware(func(res http.ResponseWriter, req *http.Request) {
		metricsHandler(res, req, storage)
	}))

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
