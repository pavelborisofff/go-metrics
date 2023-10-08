package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMemStorage_UpdateGauge(t *testing.T) {
	type testType struct {
		name      string
		gaugeName string
		values    Gauge
		expected  Gauge
	}

	tests := []testType{
		{
			name:      "UpdateGauge",
			gaugeName: "anyGauge",
			values:    Gauge(10),
			expected:  Gauge(10),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := NewMemStorage()
			storage.UpdateGauge(test.gaugeName, test.values)

			got := storage.GaugeStorage[test.gaugeName]

			if got != test.expected {
				t.Errorf("UpdateGauge() = %v, want %v", got, test.expected)
			}
		})
	}
}

func TestMemStorage_IncrementCounter(t *testing.T) {
	type testType struct {
		name        string
		counterName string
		values      Counter
		expected    Counter
	}

	tests := []testType{
		{
			name:        "UpdateCounter",
			counterName: "anyCounter",
			values:      Counter(10),
			expected:    Counter(10),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := NewMemStorage()
			storage.IncrementCounter(test.counterName, test.values)

			got := storage.CounterStorage[test.counterName]

			if got != test.expected {
				t.Errorf("UpdateCounter() = %v, want %v", got, test.expected)
			}
		})
	}
}

func TestMemStorage_SendMetric(t *testing.T) {
	storage := NewMemStorage()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("SendMetric() method = %v, want %v", r.Method, http.MethodPost)
		}

		if r.URL.Path != "/update/gauge/anyCounter/10" {
			t.Errorf("SendMetric() path = %v, want %v", r.URL.Path, "/update/gauge/anyCounter/10")
		}
	}))

	storage.SendMetric("gauge", "anyCounter", 10)
	defer server.Close()
}
