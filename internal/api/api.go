package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ulti-mate-org/logging/internal/elastic"
	"github.com/ulti-mate-org/logging/internal/metrics"
	"github.com/ulti-mate-org/logging/internal/queue"
	"github.com/ulti-mate-org/logging/internal/models"
)

// API holds dependencies like the Elasticsearch client.
type API struct {
	Elastic *elastic.ElasticClient
}

var (
	commit    string
	version   string
	buildDate string
)

// ListenAndServe starts the HTTP server on the specified address.
func (a *API) ListenAndServe(addr string) error {
	
	metrics.PrometheusInit()

	http.HandleFunc("/ingest", a.ingestHandler)
	http.HandleFunc("/health", healthCheckHandler)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/version", versionHandler)	
	
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
		var record models.LogRecord
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
			var inner models.LogRecord
			if err := json.Unmarshal([]byte(logStr), &inner); err == nil {
				record["log"] = inner
			}
		}
		
		if err := queue.Enqueue(record); err != nil {
			http.Error(w, "queueu is full", http.StatusTooManyRequests)
			return
		}

		// fmt.Println("parsed record:", record)
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

// versionHandler simply returns the git commit, build time and version tag.
func versionHandler(w http.ResponseWriter, r *http.Request) {
	
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"version":   version,
		"commit":    commit,
		"buildDate": buildDate,
	})
}

// SetBuildInfo sets the build information.
func SetBuildInfo(v, c, b string) {
	version = v
	commit = c
	buildDate = b
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
