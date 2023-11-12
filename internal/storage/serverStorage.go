package storage

import (
	"encoding/json"
	"os"
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

var (
	instance *MemStorage
	once     sync.Once
)

func NewMemStorage() *MemStorage {
	once.Do(func() {
		instance = &MemStorage{
			CounterStorage: make(map[string]Counter),
			GaugeStorage:   make(map[string]Gauge),
			mu:             &sync.Mutex{},
		}
	})

	return instance
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

func (s *MemStorage) ToFile(f string) error {
	data, err := json.MarshalIndent(s, "", "   ")
	if err != nil {
		return err
	}

	return os.WriteFile(f, data, 0644)
}

func (s *MemStorage) FromFile(f string) error {
	file, err := os.OpenFile(f, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := os.ReadFile(f)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	if err = json.Unmarshal(data, s); err != nil {
		return err
	}

	return nil
}
