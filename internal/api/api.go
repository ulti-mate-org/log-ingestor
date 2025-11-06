package api

import (
	"encoding/json"
	_ "fmt"
	"net/http"
)

// ListenAndServe starts the HTTP server on the specified address.
func ListenAndServe(addr string) error {
	http.HandleFunc("/ingest", ingestHandler)
	http.HandleFunc("/health", healthCheckHandler)
	return http.ListenAndServe(addr, nil)
}

// ingestHandler handles data ingestion requests.
func ingestHandler(w http.ResponseWriter, r *http.Request) {
	// neat neat defer, love this feature
	defer r.Body.Close()

	// need to do this since fluentbit is sending json_lines (multiple json objects in one body)
	decoder := json.NewDecoder(r.Body)
	for decoder.More() {
		var record map[string]any
		if err := decoder.Decode(&record); err != nil {
			// fmt.Println("error decoding record:", err)
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		if logStr, ok := record["log"].(string); ok {
			var inner map[string]any
			if err := json.Unmarshal([]byte(logStr), &inner); err == nil {
				record["log"] = inner
			}
		}

		// fmt.Println("parsed record:", record)
		// TODO: send record to elasticsearch
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// healthCheckHandler handles health check requests.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
