package handlers

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/pavelborisofff/go-metrics/internal/storage"
)

//go:embed templates/metrics.html
var htmlMetrics string
var s = storage.NewMemStorage()

func MainHandler(res http.ResponseWriter, _ *http.Request) {
	tmpl := template.Must(template.New("metrics").Parse(htmlMetrics))
	err := tmpl.Execute(res, s)

	if err != nil {
		log.Println(err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

}

func UpdateHandler(res http.ResponseWriter, req *http.Request) {
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

		s.IncrementCounter(metricName, storage.Counter(v))
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

		s.UpdateGauge(metricName, storage.Gauge(v))
		msg := fmt.Sprintf("Gauge %s updated to %f", metricName, v)
		log.Println(msg)

	default:
		msg := fmt.Sprintf("Bad metric's type: %s", metricType)
		log.Println(msg)
		http.Error(res, msg, http.StatusBadRequest)
	}

	res.WriteHeader(http.StatusOK)
}

func UpdateJSONHandler(res http.ResponseWriter, req *http.Request) {
	var m storage.Metrics
	var b bytes.Buffer

	_, err := b.ReadFrom(req.Body)
	if err != nil {
		msg := fmt.Sprintf("Error: %s", err)
		log.Println(msg)
		http.Error(res, msg, http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(b.Bytes(), &m)
	if err != nil {
		msg := fmt.Sprintf("Error: %s", err)
		log.Println(msg)
		http.Error(res, msg, http.StatusBadRequest)
		return
	}

	switch m.MType {
	case storage.CounterType:
		if m.Delta == nil {
			msg := fmt.Sprintf("Bad Counter's value: %s", m.ID)
			log.Println(msg)
			http.Error(res, msg, http.StatusBadRequest)
			return
		}

		s.IncrementCounter(m.ID, storage.Counter(*m.Delta))
		msg := fmt.Sprintf("Counter %s shanged to %d", m.ID, *m.Delta)
		log.Println(msg)
		res.WriteHeader(http.StatusOK)
	case storage.GaugeType:
		if m.Value == nil {
			msg := fmt.Sprintf("Bad Gauge's value: %s", m.ID)
			log.Println(msg)
			http.Error(res, msg, http.StatusBadRequest)
			return
		}

		s.UpdateGauge(m.ID, storage.Gauge(*m.Value))
		msg := fmt.Sprintf("Gauge %s updated to %f", m.ID, *m.Value)
		log.Println(msg)
		res.WriteHeader(http.StatusOK)
	default:
		msg := fmt.Sprintf("Bad metric's type: %s", m.MType)
		log.Println(msg)
		http.Error(res, msg, http.StatusBadRequest)
	}
}

func MetricsHandler(res http.ResponseWriter, _ *http.Request) {
	data, err := json.Marshal(s)
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

func ValueHandler(res http.ResponseWriter, req *http.Request) {
	metricType := chi.URLParam(req, "metric-type")
	metricName := chi.URLParam(req, "metric-name")

	if metricName == "RandomValue" {
		fmt.Sprintf("HANDLER:::RandomValue: %s", metricType)
	}

	switch metricType {
	case "counter":
		if v, ok := s.CounterStorage[metricName]; ok {
			_, err := io.WriteString(res, fmt.Sprintf("%v", v))
			if err != nil {
				return
			}
		} else {
			http.Error(res, "Not found", http.StatusNotFound)
		}
	case "gauge":
		if v, ok := s.GaugeStorage[metricName]; ok {
			_, err := io.WriteString(res, fmt.Sprintf("%v", v))
			if err != nil {
				return
			}
		} else {
			http.Error(res, "Not found", http.StatusNotFound)
		}
	default:
		http.Error(res, "Bad metric's type", http.StatusBadRequest)
	}
}

func ValueJSONHandler(res http.ResponseWriter, req *http.Request) {
	var m storage.Metrics
	var b bytes.Buffer

	_, err := b.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(b.Bytes(), &m)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	switch m.MType {
	case storage.CounterType:
		if v, ok := s.CounterStorage[m.ID]; ok {
			m.Delta = new(int64)
			*m.Delta = int64(v)
		} else {
			http.Error(res, "Not found", http.StatusNotFound)
		}
	case storage.GaugeType:
		if v, ok := s.GaugeStorage[m.ID]; ok {
			m.Value = new(float64)
			*m.Value = float64(v)
		} else {
			http.Error(res, "Not found", http.StatusNotFound)
		}
	default:
		http.Error(res, "Bad metric's type", http.StatusBadRequest)
	}

	resJSON, err := json.Marshal(m)
	if err != nil {
		msg := fmt.Sprintf("Error: %s", err)
		log.Println(msg)
		http.Error(res, msg, http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(resJSON)
}
