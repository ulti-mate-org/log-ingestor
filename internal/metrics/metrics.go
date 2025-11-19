package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
)

var (
	LogsIngested = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "app_logs_ingested_total",
			Help: "Total number of logs ingested by the application.",
		},
	)

    IngestRequests = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "ingest_requests_total",
            Help: "Number of /ingest requests received",
        },
    )

    IngestErrors = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "ingest_errors_total",
            Help: "Number of ingestor failures.",
        },
    )

    IngestLatency = prometheus.NewHistogram(
        prometheus.HistogramOpts{
            Name:    "ingest_latency_seconds",
            Help:    "Latency of ingestion requests.",
            Buckets: prometheus.DefBuckets,
        },
    )

    QueueSize = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "ingest_queue_size",
            Help: "Current number of items in the ingest queue.",
        },
    )
)

// PrometehusInit initializes Prometheus metrics and registers them.
func PrometehusInit() {
	prometheus.MustRegister(LogsIngested, IngestRequests, IngestErrors, IngestLatency, QueueSize)
}

// Collect collects and returns application metrics as a string.
// func Collect() (string, error) {
//	  li := LogsIngested.CounterValue()
//	  metrics := fmt.Sprintf("app_logs_ingested_total %f\n", li)
//    return metrics, nil
// }
