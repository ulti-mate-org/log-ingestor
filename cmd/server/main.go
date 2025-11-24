package main

import (
	"fmt"
	"context"
	"time"

	"github.com/ulti-mate-org/logging/internal/models"
	"github.com/ulti-mate-org/logging/internal/api"
	"github.com/ulti-mate-org/logging/internal/elastic"
	"github.com/ulti-mate-org/logging/internal/queue"
)

// Build information - to be set during build time
var (
    Version   = "dev"
    Commit    = "none"
    BuildDate = "unknown"
)

// TODO: tide this to env / config aswell
var (
	workers = 4
	capacity = 10000
	maxBatch = 1000
	flushInterval = 150 * time.Millisecond
)

func main() {
	fmt.Println("logging service starting...")
	// TODO: migrate to config file
	// TODO: change the logs index to a daily index
	elasticClient, err := elastic.NewElasticClient([]string{"https://elasticsearch:9200"}, "logs-date")
	if err != nil {
		fmt.Println("error creating Elasticsearch client:", err)
		return
	}

	indexFunc := func(ctx context.Context, batch []models.LogRecord) error {
		
		docs := make([]map[string]any, len(batch))
		for i, r := range batch {
			docs[i] = map[string]any(r)
		}
		return elasticClient.BulkIndex(ctx, docs)
 	}

	// we can use this wrapper to avoid different signatures
	// indexFuncTyped := func(ctx context.Context, batch []map[string]any) error {
	// 		return indexFunc(ctx, batch)
	// }
	
	// since the queue expects a func(context.Context, []models.LogRecord) error
	// here i provide a simple adapter:
	// adapter := func(ctx context.Context, batch []map[string]any) error {
	//		// return indexFuncTyped(ctx, batch)
	//		return indexFunc(ctx, batch)
	// }
	
	queue.Init(indexFunc, workers, capacity, maxBatch, flushInterval)
	

	a := &api.API{
		Elastic: elasticClient,
	}
	
	api.SetBuildInfo(Commit, Version, BuildDate)

	a.ListenAndServe(":8080")
	fmt.Println("logging service stopped.")
}
