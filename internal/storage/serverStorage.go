package storage

import (
	"encoding/json"
	"os"
	"sync"
)

type Counter uint64
type Gauge float64

type MemStorage struct {
	CounterStorage map[string]Counter `json:"counter"`
	GaugeStorage   map[string]Gauge   `json:"gauge"`
	Mu             *sync.Mutex        `json:"-"`
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
	ms       *MemStorage
	msInited sync.Once
)

func NewMemStorage() *MemStorage {
	msInited.Do(func() {
		ms = &MemStorage{
			CounterStorage: make(map[string]Counter),
			GaugeStorage:   make(map[string]Gauge),
			Mu:             &sync.Mutex{},
		}
	})

	return ms
}

func (s *MemStorage) UpdateGauge(name string, value Gauge) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	s.GaugeStorage[name] = value
}

func (s *MemStorage) IncrementCounter(name string, value Counter) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

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
