package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ulti-mate-org/logging/internal/elastic"
)

// API holds dependencies like the Elasticsearch client.
type API struct {
	Elastic *elastic.ElasticClient
}

// ListenAndServe starts the HTTP server on the specified address.
func (a *API) ListenAndServe(addr string) error {
	http.HandleFunc("/ingest", a.ingestHandler)
	http.HandleFunc("/health", healthCheckHandler)
	return http.ListenAndServe(addr, nil)
}

// ingestHandler handles data ingestion requests.
func (a *API) ingestHandler(w http.ResponseWriter, r *http.Request) {
	// neat neat defer, love this feature
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// need to do this since fluentbit is sending json_lines (multiple json objects in one body)
	decoder := json.NewDecoder(r.Body)
	count := 0

	for decoder.More() {
		var record map[string]any
		if err := decoder.Decode(&record); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("error decoding record:", err)
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
			http.Error(w, "failed to index document", http.StatusInternalServerError)
			return
		}

		count++
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("ok (%d docs indexed)", count)))
}

// healthCheckHandler handles health check requests.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
