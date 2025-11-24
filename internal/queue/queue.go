package queue

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ulti-mate-org/logging/internal/metrics"
	"github.com/ulti-mate-org/logging/internal/models"
)

var (
	QUEUE_CAPACITY = 10000
	WORKER_COUNT   = 4
	MAX_BATCH_SIZE = 1000
	FLUSH_INTERVAL = 150 * time.Millisecond

	// ErrQueueFull returned when the queue is full and the record cannot be enqueued
	ErrQueueFull = errors.New("ingest queue is full")
)

var (
	logQueue chan models.LogRecord
	shutdownCancel context.CancelFunc
)

// Init starts the queue and worker goroutines.
// indexFunc: function called with a batch to perform the actual bulk indexing.
// workers, capacity, maxBatch, flushInterval override defaults if >0.
func Init(indexFunc func(ctx context.Context, batch []models.LogRecord) error, workers, capacity, maxBatch int, flushInterval time.Duration) {
	if capacity > 0 {
		QUEUE_CAPACITY = capacity
	}
	if workers > 0 {
		WORKER_COUNT = workers
	}
	if maxBatch > 0 {
		MAX_BATCH_SIZE = maxBatch
	}
	if flushInterval > 0 {
		FLUSH_INTERVAL = flushInterval
	}

	logQueue = make(chan models.LogRecord, QUEUE_CAPACITY)

	ctx, cancel := context.WithCancel(context.Background())
	shutdownCancel = cancel

	// start background workers
	for i := 0; i < WORKER_COUNT; i++ {
		go worker(ctx, indexFunc, i)
	}

	// start a small ticker to publish queue size metric periodically
	go func() {
		t := time.NewTicker(250 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if logQueue != nil {
					metrics.QueueSize.Set(float64(len(logQueue)))
				}
			}
		}
	}()
}

// Enqueue attempts to push a record into the queue.
// Non-blocking: returns ErrQueueFull if the queue is full.
func Enqueue(rec models.LogRecord) error {
	select {
	case logQueue <- rec:
		metrics.QueueSize.Set(float64(len(logQueue)))
		return nil
	default:
		metrics.IngestErrors.Inc()
		metrics.QueueOverflows.Inc() 
		return ErrQueueFull
	}
}

// Shutdown stops workers and waits for background goroutines to finish until ctx expires.
// It will attempt to drain the queue before returning.
func Shutdown(ctx context.Context) error {
	if shutdownCancel != nil {
		shutdownCancel()
	}

	// wait for queue to drain or ctx timeout
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if len(logQueue) == 0 {
				return nil
			}
		}
	}
}

// worker consumes from logQueue, builds batches, and calls indexFunc.
// It performs retries with exponential backoff for failed batches and reports metrics.
func worker(ctx context.Context, indexFunc func(ctx context.Context, batch []models.LogRecord) error, id int) {
	batch := make([]models.LogRecord, 0, MAX_BATCH_SIZE)
	flushTimer := time.NewTimer(FLUSH_INTERVAL)
	defer flushTimer.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		toSend := make([]models.LogRecord, len(batch))
		copy(toSend, batch)
		batch = batch[:0]

		var attempt int
		backoff := 200 * time.Millisecond

		metrics.BulkBatchSize.Observe(float64(len(toSend)))

		for {
			attempt++
			metrics.BulkRequests.Inc()
			start := time.Now()

			callCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			err := indexFunc(callCtx, toSend)
			cancel()

			metrics.BulkLatency.Observe(time.Since(start).Seconds())

			if err == nil {
				return
			}

			metrics.BulkFailures.Inc()

			// TODO: decide whevDther to retry
			if attempt >= 5 {
				fmt.Printf("[queue] worker %d: dropping batch after %d attempts: %v\n", id, attempt, err)
				return
			}

			metrics.BulkRetries.Inc()

			sleep := backoff + time.Duration((attempt*23)%50)*time.Millisecond
			select {
			case <-time.After(sleep):
				backoff *= 2
				continue
			case <-ctx.Done():
				return
			}
		}
	}

	for {
		// if we have a full batch, flush immediately (fast path)
		if len(batch) >= MAX_BATCH_SIZE {
			flush()
			if !flushTimer.Stop() {
				select {
				case <-flushTimer.C:
				default:
				}
			}
			flushTimer.Reset(FLUSH_INTERVAL)
			continue
		}

		select {
		case <-ctx.Done():
			
			flush()
			return
		case rec := <-logQueue:
			batch = append(batch, rec)
			// if this is first element, ensure timer is running
			if len(batch) == 1 {
				if !flushTimer.Stop() {
					select {
					case <-flushTimer.C:
					default:
					}
				}
				flushTimer.Reset(FLUSH_INTERVAL)
			}
		case <-flushTimer.C:
			flush()
			flushTimer.Reset(FLUSH_INTERVAL)
		}
	}
}

