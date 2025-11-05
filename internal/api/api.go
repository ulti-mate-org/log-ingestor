package api

import (
	"fmt"
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
	// TODO: implement log ingestion logic and send to elasticsearch
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Data ingested successfully"))
}

// healthCheckHandler handles health check requests.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
