package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ulti-mate-org/logging/internal/elastic"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ulti-mate-org/logging/internal/metrics"
)

// API holds dependencies like the Elasticsearch client.
type API struct {
	Elastic *elastic.ElasticClient
}

// ListenAndServe starts the HTTP server on the specified address.
func (a *API) ListenAndServe(addr string) error {
	
	metrics.PrometehusInit()

	http.HandleFunc("/ingest", a.ingestHandler)
	http.HandleFunc("/health", healthCheckHandler)
	http.Handle("/metrics", promhttp.Handler())

	return http.ListenAndServe(addr, nil)
}

// ingestHandler handles data ingestion requests.
func (a *API) ingestHandler(w http.ResponseWriter, r *http.Request) {
	// use prometheus for metrics calculation
	timer := prometheus.NewTimer(metrics.IngestLatency)
    defer timer.ObserveDuration()

	// neat neat defer, love this feature
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics.IngestRequests.Inc()

	decoder := json.NewDecoder(r.Body)
	count := 0

	for decoder.More() {
		var record map[string]any
		if err := decoder.Decode(&record); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("error decoding record:", err)
			metrics.IngestErrors.Inc()
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// handle nested json since log contains data as a string
		if logStr, ok := record["log"].(string); ok {
			var inner map[string]any
			if err := json.Unmarshal([]byte(logStr), &inner); err == nil {
				record["log"] = inner
			}
		}

		// fmt.Println("parsed record:", record)

		// index the record
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		err := a.Elastic.IndexDocument(ctx, record)
		cancel()
		if err != nil {
			fmt.Println("error indexing record:", err)
			metrics.IngestErrors.Inc()
			http.Error(w, "failed to index document", http.StatusInternalServerError)
			return
		}

		count++
	}

	metrics.LogsIngested.Add(float64(count))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("ok (%d docs indexed)", count)))
}

// healthCheckHandler handles health check requests.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// metricsHandler handles metrics requests.
// func metricsHandler(w http.ResponseWriter, r *http.Request) {
	
//	if r.Method != http.MethodGet {
//		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
//	}
//
//	m, err := metrics.Collect()
//	if err != nil {
//		fmt.Println("error collecting metrics:", err)
//		http.Error(w, "failed to collect metrics", http.StatusInternalServerError)
//		return
//	}
//
//	w.Write([]byte(m))
// }
