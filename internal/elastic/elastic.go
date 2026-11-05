package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// ElasticClient wraps the Elasticsearch client.
type ElasticClient struct {
	Client *elasticsearch.Client
	Index  string
}

// Config holds the Elasticsearch configuration.
type Config struct {
	Addresses []string
	APIKey    string
	CAPath    string
	Index     string
}

// NewElasticClient creates a new Elasticsearch client.
func NewElasticClient(addresses []string, index string) (*ElasticClient, error) {
	caCertPath := os.Getenv("ELASTIC_CA_CERT")
	apiKey := os.Getenv("ELASTIC_API_KEY")

	fmt.Println("CA Cert Path: ", caCertPath)
	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA cert: %w", err)
	}

	cfg := elasticsearch.Config{
		Addresses: addresses,
		APIKey:    apiKey,
		CACert:    []byte(caCert), // pass the CA contents as bytes
	}
	fmt.Println("config: ", cfg)
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
		return fmt.Errorf("encode error: %w", err)
	}

	req := esapi.IndexRequest{
		Index:   ec.Index,
		Body:    &buf,
		Refresh: "true",
	}

	res, err := req.Do(ctx, ec.Client)
	if err != nil {
		// consider 429 (too many req) and 503 (cluster problem) errors
		if res != nil {
			if res.StatusCode == 429 {
				return fmt.Errorf("elasticsearch rate limit exceeded: %w", err)
			} else if res.StatusCode == 503 {
				return fmt.Errorf("elasticsearch service unavailable: %w", err)
			}
		}
		return fmt.Errorf("request error: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch error: %s", res.String())
	}

	return nil
}
