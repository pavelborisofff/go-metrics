package handlers

import (
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

func MainHandler(res http.ResponseWriter, _ *http.Request, storage *storage.MemStorage) {
	tmpl := template.Must(template.New("metrics").Parse(htmlMetrics))
	err := tmpl.Execute(res, storage)

	if err != nil {
		log.Println(err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}

}

func UpdateHandler(res http.ResponseWriter, req *http.Request, s *storage.MemStorage) {
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

func MetricsHandler(res http.ResponseWriter, _ *http.Request, s *storage.MemStorage) {
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

func ValueHandler(res http.ResponseWriter, req *http.Request, s *storage.MemStorage) {
	metricType := chi.URLParam(req, "metric-type")
	metricName := chi.URLParam(req, "metric-name")

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
