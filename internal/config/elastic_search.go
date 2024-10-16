package config

import (
	"fmt"
	"log"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/joho/godotenv"
)

var (
	ESClient *elasticsearch.Client
)

// InitElasticsearch initializes the Elasticsearch client
func InitElasticsearch() error {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	// Get Elasticsearch configuration from environment variables
	elasticEndpoint := os.Getenv("ELASTIC_SEARCH_ENDPOINT")
	elasticAPIKey := os.Getenv("ELASTIC_SEARCH_API_KEY")

	// Create the Elasticsearch client configuration
	cfg := elasticsearch.Config{
		Addresses: []string{elasticEndpoint},
		APIKey:    elasticAPIKey,
	}

	// Create the Elasticsearch client
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("error creating the Elasticsearch client: %w", err)
	}

	// Check if the connection is successful
	res, err := client.Info()
	if err != nil {
		return fmt.Errorf("error connecting to Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	ESClient = client
	log.Println("Connected to Elasticsearch successfully!")
	return nil
}

// GetESClient returns the Elasticsearch client
func GetESClient() *elasticsearch.Client {
	return ESClient
}
