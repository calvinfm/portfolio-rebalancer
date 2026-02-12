package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"portfolio-rebalancer/internal/models"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/google/uuid"
)

var esClient *elasticsearch.Client

// InitElastic initializes elasticsearch connection with retry logic
func InitElastic() error {
	cfg := elasticsearch.Config{
		Addresses: []string{
			os.Getenv("ELASTICSEARCH_URL"),
		},
	}

	var client *elasticsearch.Client
	var err error

	for i := 1; i <= 5; i++ {
		client, err = elasticsearch.NewClient(cfg)
		if err != nil {
			log.Printf("Failed to create client: %v", err)
		} else {
			_, err = client.Info()
			if err == nil {
				log.Println("Connected to Elasticsearch")
				esClient = client
				return nil
			}
			log.Printf("Client created, but ES not ready: %v", err)
		}

		log.Printf("Retrying connection to Elasticsearch... (%d/5)", i)
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("failed to connect to Elasticsearch after retries: %w", err)
}

func SavePortfolio(ctx context.Context, p models.Portfolio) error {
	body, err := json.Marshal(p)
	if err != nil {
		return err
	}

	res, err := esClient.Index("portfolios", bytes.NewReader(body), esClient.Index.WithDocumentID(p.UserID))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error saving portfolio: %s", res.String())
	}

	log.Printf("Portfolio saved for user %s", p.UserID)
	return nil
}

func GetPortfolio(ctx context.Context, userID string) (*models.Portfolio, error) {
	res, err := esClient.Get("portfolios", userID)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("user not found")
	}

	var esResp struct {
		Source models.Portfolio `json:"_source"`
	}

	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, err
	}

	return &esResp.Source, nil
}

func SaveRebalanceTransaction(ctx context.Context, p []models.RebalanceTransaction) error {
	for _, v := range p {
		v.Id = uuid.New().String()

		body, err := json.Marshal(v)
		if err != nil {
			return err
		}

		res, err := esClient.Index("rebalance_transactions", bytes.NewReader(body), esClient.Index.WithDocumentID(v.Id))
		if err != nil {
			return err
		}

		if res.IsError() {
			return fmt.Errorf("error saving rebalance transactions: %s", res.String())
		}
		res.Body.Close()
	}

	log.Printf("Rebalance transactions saved for users %+v", p)
	return nil
}

func SaveRebalanceTransactionBulk(ctx context.Context, transactions []models.RebalanceTransaction) error {
	if len(transactions) == 0 {
		return nil
	}

	var buf bytes.Buffer
	for _, v := range transactions {
		if v.Id == "" {
			v.Id = uuid.New().String()
		}

		// 1. Action metadata line
		meta := []byte(fmt.Sprintf(`{ "index" : { "_index" : "rebalance_transactions", "_id" : "%s" } }%s`, v.Id, "\n"))

		// 2. Document data line
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		data = append(data, "\n"...)

		buf.Write(meta)
		buf.Write(data)
	}

	// 3. Send the single bulk request
	res, err := esClient.Bulk(
		bytes.NewReader(buf.Bytes()),
		esClient.Bulk.WithContext(ctx),
		esClient.Bulk.WithRefresh("wait_for"), // ⭐ Forces the function to wait until data is searchable
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk index error: %s", res.String())
	}

	log.Printf("Successfully bulk indexed %d transactions", len(transactions))
	return nil
}

func GetRebalanceTransaction(ctx context.Context, userID string) (*[]models.RebalanceTransaction, error) {
	res, err := esClient.Search(esClient.Search.WithContext(ctx),
		esClient.Search.WithIndex("rebalance_transactions"),
		esClient.Search.WithPretty())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("user not found")
	}

	var esResp struct {
		Source []models.RebalanceTransaction `json:"_source"`
	}

	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, err
	}

	return &esResp.Source, nil
}

func GetTransactionsByUserID(ctx context.Context, userID string) ([]models.RebalanceTransaction, error) {
	// 1. Define the search query
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"user_id": userID, // Exact match search
			},
		},
	}

	// 2. Encode to JSON buffer
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, err
	}

	// 3. Execute Search
	res, err := esClient.Search(
		esClient.Search.WithContext(ctx),
		esClient.Search.WithIndex("rebalance_transactions"),
		esClient.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search error: %s", res.String())
	}

	// 4. Parse the Response
	var result struct {
		Hits struct {
			Hits []struct {
				Source models.RebalanceTransaction `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	// 5. Extract the transactions from the "hits"
	transactions := make([]models.RebalanceTransaction, len(result.Hits.Hits))
	for i, hit := range result.Hits.Hits {
		transactions[i] = hit.Source
	}

	return transactions, nil
}
