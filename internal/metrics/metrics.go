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

	QueueOverflows = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ingest_queue_overflows_total",
			Help: "Number of times an enqueue attempt failed because the queue was full.",
		},
	)

	BulkRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ingest_bulk_requests_total",
			Help: "Number of bulk requests sent to Elasticsearch.",
		},
	)

	BulkFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ingest_bulk_failures_total",
			Help: "Total number of bulk failures to Elasticsearch.",
		},
	)

	BulkRetries = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ingest_bulk_retries_total",
			Help: "Total number of bulk retries performed.",
		},
	)

	BulkLatency = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "ingest_bulk_latency_seconds",
			Help:    "Latency of bulk requests to Elasticsearch.",
			Buckets: prometheus.ExponentialBuckets(0.05, 2, 10), // 50ms..~25s
		},
	)

	BulkBatchSize = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "ingest_bulk_batch_size",
			Help:    "Number of documents in each bulk request.",
			Buckets: prometheus.ExponentialBuckets(50, 2, 6), // 50,100,200,400,800,1600
		},
	)
)

// PrometheusInit initializes Prometheus metrics and registers them.
func PrometheusInit() {
	prometheus.MustRegister(
		LogsIngested,
		IngestRequests, 
		IngestErrors, 
		IngestLatency, 
		QueueSize,
		QueueOverflows,
		BulkFailures,
		BulkRetries,
		BulkLatency,
		BulkBatchSize,
		BulkRequests,
	)
}

// Collect collects and returns application metrics as a string.
// func Collect() (string, error) {
//	  li := LogsIngested.CounterValue()
//	  metrics := fmt.Sprintf("app_logs_ingested_total %f\n", li)
//    return metrics, nil
// }
