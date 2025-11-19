package main

import (
	"fmt"
	"github.com/ulti-mate-org/logging/internal/api"
	"github.com/ulti-mate-org/logging/internal/elastic"
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
	
	api := &api.API{
		Elastic: elasticClient,
	}

	api.ListenAndServe(":8080")
	fmt.Println("logging service stopped.")
}
