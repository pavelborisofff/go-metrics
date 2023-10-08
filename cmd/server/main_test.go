package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMainHandler(t *testing.T) {
	type testType struct {
		name         string
		expectedCode int
		expectedBody string
	}

	tests := []testType{
		{
			name:         "Receive 400 with request to root",
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			mainHandler(w, r)

			assert.Equal(t, test.expectedCode, w.Code, "Status code mismatch")

			if test.expectedBody != "" {
				assert.JSONEq(t, test.expectedBody, w.Body.String(), "Response body mismatch")
			}
		})
	}
}

func TestUpdateHandler(t *testing.T) {
	type testType struct {
		name         string
		method       string
		requestURL   string
		expectedCode int
		expectedBody string
	}

	tests := []testType{
		{
			name:         "Receive 405 if method is not POST",
			method:       http.MethodGet,
			requestURL:   "/update/",
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: "",
		},
		{
			name:         "Update Counter",
			method:       http.MethodPost,
			requestURL:   "/update/counter/anyCounter/5",
			expectedCode: http.StatusOK,
			expectedBody: "",
		},
		{
			name:         "Update Counter with invalid value",
			method:       http.MethodPost,
			requestURL:   "/update/counter/anyCounter/invalid",
			expectedCode: http.StatusBadRequest,
			expectedBody: "Bad Counter's value: invalid\n",
		},
		{
			name:         "Update Gauge",
			method:       http.MethodPost,
			requestURL:   "/update/gauge/anyGauge/123.123123",
			expectedCode: http.StatusOK,
			expectedBody: "",
		},
		{
			name:         "Update Gauge with invalid value",
			method:       http.MethodPost,
			requestURL:   "/update/gauge/anyGauge/invalid",
			expectedCode: http.StatusBadRequest,
			expectedBody: "Bad metric's value: anyGauge invalid\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := NewMemStorage()
			r := httptest.NewRequest(test.method, test.requestURL, nil)
			w := httptest.NewRecorder()

			updateHandler(w, r, storage)

			assert.Equal(t, test.expectedCode, w.Code, "Status code mismatch")

			if test.expectedBody != "" {
				assert.Equal(t, test.expectedBody, w.Body.String(), "Response body mismatch")
			}
		})
	}
}

func TestMetricsHandler(t *testing.T) {
	type testType struct {
		name         string
		method       string
		requestURL   string
		expectedCode int
		expectedBody string
	}

	tests := []testType{
		{
			name:         "Receive 405 if method is not GET",
			method:       http.MethodPost,
			requestURL:   "/metrics/",
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: "",
		},
		{
			name:         "Get Metrics",
			method:       http.MethodGet,
			requestURL:   "/metrics/",
			expectedCode: http.StatusOK,
			expectedBody: `{"counter":{},"gauge":{}}`,
		},
		{
			name:         "Get Metrics with data",
			method:       http.MethodGet,
			requestURL:   "/metrics/",
			expectedCode: http.StatusOK,
			expectedBody: `{"counter":{},"gauge":{"anygauge":123.123123}}`,
		},
	}

	storage := NewMemStorage()

	for _, test := range tests[:len(tests)-1] {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(test.method, test.requestURL, nil)
			w := httptest.NewRecorder()

			metricsHandler(w, r, storage)

			assert.Equal(t, test.expectedCode, w.Code, "Status code mismatch")

			if test.expectedBody != "" {
				assert.Equal(t, test.expectedBody, w.Body.String(), "Response body mismatch")
			}
		})
	}

	t.Run(tests[len(tests)-1].name, func(t *testing.T) {
		storage.UpdateGauge("anygauge", 123.123123)
		r := httptest.NewRequest(tests[len(tests)-1].method, tests[len(tests)-1].requestURL, nil)
		w := httptest.NewRecorder()

		metricsHandler(w, r, storage)

		assert.Equal(t, tests[len(tests)-1].expectedCode, w.Code, "Status code mismatch")

		if tests[len(tests)-1].expectedBody != "" {
			assert.Equal(t, tests[len(tests)-1].expectedBody, w.Body.String(), "Response body mismatch")
		}
	})
}
