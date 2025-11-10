package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// ElasticClient wraps the Elasticsearch client.
type ElasticClient struct {
	Client *elasticsearch.Client
	Index  string
}

// NewElasticClient creates a new Elasticsearch client.
func NewElasticClient(addresses []string, index string) (*ElasticClient, error) {
	cfg := elasticsearch.Config{
		Addresses: addresses,
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating Elasticsearch client: %w", err)
	}
	return &ElasticClient{Client: es, Index: index}, nil
}

// IndexDocument indexes a document in Elasticsearch.
func (ec *ElasticClient) IndexDocument(ctx context.Context, doc map[string]any) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(doc); err != nil {
		return fmt.Errorf("error encoding document to JSON: %w", err)
	}

	req := esapi.IndexRequest{
		Index: ec.Index,
		// TODO: use a custom id (DocumentID: ,.,.,.)
		Body:    &buf,
		Refresh: "true",
	}

	res, err := req.Do(ctx, ec.Client)
	if err != nil {
		return fmt.Errorf("error indexing document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error response from Elasticsearch: %s", res.String())
	}

	fmt.Printf("Document indexed successfully: %s\n", res.String())
	return nil
}
