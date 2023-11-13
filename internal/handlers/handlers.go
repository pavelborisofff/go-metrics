package handlers

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/pavelborisofff/go-metrics/internal/db"
	"html/template"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/pavelborisofff/go-metrics/internal/logger"
	"github.com/pavelborisofff/go-metrics/internal/storage"
)

//go:embed templates/metrics.html
var htmlMetrics string

var (
	s   = storage.NewMemStorage()
	log = logger.GetLogger()
)

func MainHandler(res http.ResponseWriter, _ *http.Request) {
	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	tmpl := template.Must(template.New("metrics").Parse(htmlMetrics))
	err := tmpl.Execute(res, s)

	if err != nil {
		log.Error("Error execute template", zap.Error(err))
		http.Error(res, "Error execute template", http.StatusInternalServerError)
		return
	}
}

func PingHandler(res http.ResponseWriter, _ *http.Request) {
	if db.DB == nil {
		log.Debug("DB is nil")
		http.Error(res, "DB is nil", http.StatusInternalServerError)
		return
	}

	if err := db.DB.Ping(context.Background()); err != nil {
		http.Error(res, "Error connecting to DB", http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
	res.Write([]byte("Ping successful"))
}

func UpdateHandler(res http.ResponseWriter, req *http.Request) {
	metricType := chi.URLParam(req, "metric-type")
	metricName := chi.URLParam(req, "metric-name")
	metricValue := chi.URLParam(req, "metric-value")

	switch metricType {
	case storage.CounterType:
		v, err := strconv.ParseUint(metricValue, 10, 64)

		if err != nil {
			msg := fmt.Sprintf("Bad Counter's value: %s", metricValue)
			log.Debug(msg)
			http.Error(res, msg, http.StatusBadRequest)
			return
		}

		s.IncrementCounter(metricName, storage.Counter(v))
		log.Debug("Counter change", zap.String("name", metricName), zap.Uint64("value", v))

	case storage.GaugeType:
		v, err := strconv.ParseFloat(metricValue, 64)

		if err != nil {
			msg := fmt.Sprintf("Bad metric's value: %s %s", metricName, metricValue)
			log.Debug(msg)
			http.Error(res, msg, http.StatusBadRequest)
			return
		}

		s.UpdateGauge(metricName, storage.Gauge(v))
		log.Debug("Gauge change", zap.String("name", metricName), zap.Float64("value", v))

	default:
		msg := fmt.Sprintf("Bad metric's type: %s", metricType)
		log.Debug(msg)
		http.Error(res, msg, http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusOK)
}

func UpdateJSONHandler(res http.ResponseWriter, req *http.Request) {
	var m storage.Metrics
	var b bytes.Buffer

	_, err := b.ReadFrom(req.Body)
	if err != nil {
		msg := "Error read body"
		log.Debug("Error read body", zap.Error(err))
		http.Error(res, msg, http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(b.Bytes(), &m)
	if err != nil {
		msg := "Error unmarshal"
		log.Debug("Error unmarshal", zap.Error(err))
		http.Error(res, msg, http.StatusBadRequest)
		return
	}

	switch m.MType {
	case storage.CounterType:
		if m.Delta == nil {
			msg := fmt.Sprintf("Bad Counter's value: %s", m.ID)
			log.Debug(msg)
			http.Error(res, msg, http.StatusBadRequest)
			return
		}

		s.IncrementCounter(m.ID, storage.Counter(*m.Delta))
		msg := fmt.Sprintf("Counter %s shanged to %d", m.ID, *m.Delta)
		log.Debug(msg)
		res.WriteHeader(http.StatusOK)
	case storage.GaugeType:
		if m.Value == nil {
			msg := fmt.Sprintf("Bad Gauge's value: %s", m.ID)
			log.Debug(msg)
			http.Error(res, msg, http.StatusBadRequest)
			return
		}

		s.UpdateGauge(m.ID, storage.Gauge(*m.Value))
		msg := fmt.Sprintf("Gauge %s updated to %f", m.ID, *m.Value)
		log.Debug(msg)
		res.WriteHeader(http.StatusOK)
	default:
		msg := fmt.Sprintf("Bad metric's type: %s", m.MType)
		log.Debug(msg)
		http.Error(res, msg, http.StatusBadRequest)
		return
	}
}

func MetricsHandler(res http.ResponseWriter, _ *http.Request) {
	data, err := json.Marshal(s)
	if err != nil {
		msg := "Error marshal"
		log.Debug(msg, zap.Error(err))
		http.Error(res, msg, http.StatusInternalServerError)
		return
	}

	log.Debug("Metrics", zap.ByteString("data", data))
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)

	_, err = res.Write(data)
	if err != nil {
		msg := "Error write"
		log.Debug("Error write", zap.Error(err))
		http.Error(res, msg, http.StatusInternalServerError)
		return
	}
}

func ValueHandler(res http.ResponseWriter, req *http.Request) {
	metricType := chi.URLParam(req, "metric-type")
	metricName := chi.URLParam(req, "metric-name")

	switch metricType {
	case storage.CounterType:
		if v, ok := s.CounterStorage[metricName]; ok {
			_, err := io.WriteString(res, fmt.Sprintf("%v", v))
			if err != nil {
				msg := "Error write"
				log.Debug(msg, zap.Error(err))
				http.Error(res, msg, http.StatusInternalServerError)
				return
			}
		} else {
			msg := "Not found"
			log.Debug(msg, zap.String("name", metricName))
			http.Error(res, msg, http.StatusNotFound)
			return
		}
	case storage.GaugeType:
		if v, ok := s.GaugeStorage[metricName]; ok {
			_, err := io.WriteString(res, fmt.Sprintf("%v", v))
			if err != nil {
				msg := "Error write"
				log.Debug(msg, zap.Error(err))
				http.Error(res, msg, http.StatusInternalServerError)
				return
			}
		} else {
			msg := "Not found"
			log.Debug(msg, zap.String("name", metricName))
			http.Error(res, msg, http.StatusNotFound)
			return
		}
	default:
		msg := "Bad metric's type"
		log.Debug(msg, zap.String("type", metricType))
		http.Error(res, msg, http.StatusBadRequest)
		return
	}
}

func ValueJSONHandler(res http.ResponseWriter, req *http.Request) {
	var m storage.Metrics
	var b bytes.Buffer

	_, err := b.ReadFrom(req.Body)
	if err != nil {
		msg := "Error read body"
		log.Debug(msg, zap.Error(err))
		http.Error(res, msg, http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(b.Bytes(), &m)
	if err != nil {
		msg := "Error unmarshal"
		log.Debug(msg, zap.Error(err))
		http.Error(res, msg, http.StatusBadRequest)
		return
	}

	switch m.MType {
	case storage.CounterType:
		if v, ok := s.CounterStorage[m.ID]; ok {
			m.Delta = new(int64)
			*m.Delta = int64(v)
		} else {
			msg := "Not found"
			log.Debug(msg, zap.String("name", m.ID))
			http.Error(res, msg, http.StatusNotFound)
			return
		}
	case storage.GaugeType:
		if v, ok := s.GaugeStorage[m.ID]; ok {
			m.Value = new(float64)
			*m.Value = float64(v)
		} else {
			msg := "Not found"
			log.Debug(msg, zap.String("name", m.ID))
			http.Error(res, msg, http.StatusNotFound)
			return
		}
	default:
		msg := "Bad metric's type"
		log.Debug(msg, zap.String("type", m.MType))
		http.Error(res, msg, http.StatusBadRequest)
		return
	}

	resJSON, err := json.Marshal(m)
	if err != nil {
		msg := "Error marshal"
		log.Debug(msg, zap.Error(err))
		http.Error(res, msg, http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	_, err = res.Write(resJSON)
	if err != nil {
		return
	}
}
