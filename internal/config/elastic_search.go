package config

import (
	"fmt"
	"log"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/joho/godotenv"
)

var (
	ESClientPantries *elasticsearch.Client
	ESClientRecipes  *elasticsearch.Client
)

// InitElasticsearch initializes the Elasticsearch clients
func InitElasticsearch() error {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	// Get Elasticsearch configuration from environment variables
	elasticEndpoint := os.Getenv("ELASTIC_SEARCH_ENDPOINT")
	elasticAPIKeyPantries := os.Getenv("ELASTIC_SEARCH_API_KEY_PANTRIES")
	elasticAPIKeyRecipes := os.Getenv("ELASTIC_SEARCH_API_KEY_RECIPES")

	// Create the Elasticsearch client configuration for pantries
	cfgPantries := elasticsearch.Config{
		Addresses: []string{elasticEndpoint},
		APIKey:    elasticAPIKeyPantries,
	}

	// Create the Elasticsearch client configuration for recipes
	cfgRecipes := elasticsearch.Config{
		Addresses: []string{elasticEndpoint},
		APIKey:    elasticAPIKeyRecipes,
	}

	// Create the Elasticsearch clients
	clientPantries, err := elasticsearch.NewClient(cfgPantries)
	if err != nil {
		return fmt.Errorf("error creating the Elasticsearch client for pantries: %w", err)
	}

	clientRecipes, err := elasticsearch.NewClient(cfgRecipes)
	if err != nil {
		return fmt.Errorf("error creating the Elasticsearch client for recipes: %w", err)
	}

	// Check if the connections are successful
	resPantries, err := clientPantries.Info()
	if err != nil {
		return fmt.Errorf("error connecting to Elasticsearch for pantries: %w", err)
	}
	defer resPantries.Body.Close()

	resRecipes, err := clientRecipes.Info()
	if err != nil {
		return fmt.Errorf("error connecting to Elasticsearch for recipes: %w", err)
	}
	defer resRecipes.Body.Close()

	ESClientPantries = clientPantries
	ESClientRecipes = clientRecipes
	log.Println("Connected to Elasticsearch successfully for both pantries and recipes!")
	return nil
}

// GetESClientPantries returns the Elasticsearch client for pantries
func GetESClientPantries() *elasticsearch.Client {
	return ESClientPantries
}

// GetESClientRecipes returns the Elasticsearch client for recipes
func GetESClientRecipes() *elasticsearch.Client {
	return ESClientRecipes
}
