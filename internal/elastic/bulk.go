package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// BulkIndex indexes a batch of documents using the Elasticsearch _bulk API.
func (ec *ElasticClient) BulkIndex(ctx context.Context, batch []map[string]any) error {
	if len(batch) == 0 {
		return nil
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)

	// build NDJSON bulk payload: meta \n doc \n
	for _, doc := range batch {
		meta := map[string]map[string]any{
			"index": {"_index": ec.Index},
		}
		if err := enc.Encode(meta); err != nil {
			return fmt.Errorf("bulk encode meta: %w", err)
		}
		if err := enc.Encode(doc); err != nil {
			return fmt.Errorf("bulk encode doc: %w", err)
		}
	}

	res, err := ec.Client.Bulk(bytes.NewReader(buf.Bytes()), ec.Client.Bulk.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("bulk request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		
		body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return fmt.Errorf("bulk request status %d: %s", res.StatusCode, string(body))
	}

	var resp struct {
		Errors bool `json:"errors"`
		Items  []map[string]map[string]any `json:"items"`
	}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return fmt.Errorf("failed to decode bulk response: %w", err)
	}

	if !resp.Errors {
		return nil
	}

	var sb bytes.Buffer
	sb.WriteString("bulk response contained errors; sample:\n")
	n := 0
	for _, item := range resp.Items {
		// item has an action key (index/create/...), we only expect index
		for action, info := range item {
			if infoErr, ok := info["error"]; ok && infoErr != nil {
				b, _ := json.Marshal(map[string]any{
					"action": action,
					"status": info["status"],
					"error":  infoErr,
				})
				sb.Write(b)
				sb.WriteByte('\n')
				n++
				if n >= 5 {
					break
				}
			}
		}
		if n >= 5 {
			break
		}
	}
	return errors.New(sb.String())
}

