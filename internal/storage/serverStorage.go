package storage

import (
	"sync"
)

type Gauge float64
type Counter uint64

type MemStorage struct {
	CounterStorage map[string]Counter `json:"counter"`
	GaugeStorage   map[string]Gauge   `json:"gauge"`
	mu             *sync.Mutex
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

const (
	CounterType = "counter"
	GaugeType   = "gauge"
)

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
