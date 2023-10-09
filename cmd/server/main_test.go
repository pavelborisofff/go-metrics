package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestMainHandler(t *testing.T) {
	type testType struct {
		name         string
		expectedCode int
		expectedBody string
	}

	tests := []testType{
		{
			name:         "Receive 200 with request to root",
			expectedCode: http.StatusOK,
			expectedBody: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := NewMemStorage()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			mainHandler(w, r, storage)

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
			name:         "Receive 404 if method is not POST",
			method:       http.MethodGet,
			requestURL:   "/update/",
			expectedCode: http.StatusNotFound,
			expectedBody: "404 page not found\n",
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

	storage := NewMemStorage()
	r := chi.NewRouter()

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

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, test := range tests {
		fmt.Println(test.name)
		fmt.Println(test.requestURL)
		res, body := testRequest(t, ts, test.method, test.requestURL)
		assert.Equal(t, test.expectedCode, res.StatusCode)
		assert.Equal(t, test.expectedBody, body)

		res.Body.Close()
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
			requestURL:   "/metrics",
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: "",
		},
		{
			name:         "Get Metrics",
			method:       http.MethodGet,
			requestURL:   "/metrics",
			expectedCode: http.StatusOK,
			expectedBody: `{"counter":{},"gauge":{}}`,
		},
		{
			name:         "Get Metrics with data",
			method:       http.MethodGet,
			requestURL:   "/metrics",
			expectedCode: http.StatusOK,
			expectedBody: `{"counter":{},"gauge":{"anygauge":123.123123}}`,
		},
	}

	storage := NewMemStorage()
	r := chi.NewRouter()

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

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, test := range tests[:len(tests)-1] {
		fmt.Println(test.name)
		fmt.Println(test.requestURL)
		res, body := testRequest(t, ts, test.method, test.requestURL)
		assert.Equal(t, test.expectedCode, res.StatusCode)
		assert.Equal(t, test.expectedBody, body)

		res.Body.Close()
	}

	for _, test := range tests[len(tests)-1:] {
		storage.UpdateGauge("anygauge", 123.123123)
		fmt.Println(test.name)
		fmt.Println(test.requestURL)
		res, body := testRequest(t, ts, test.method, test.requestURL)
		assert.Equal(t, test.expectedCode, res.StatusCode)
		assert.Equal(t, test.expectedBody, body)

		res.Body.Close()
	}

}
