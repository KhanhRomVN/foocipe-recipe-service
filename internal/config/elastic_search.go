package config

import (
	"fmt"
	"log"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/joho/godotenv"
)

var (
	ESClientRecipes     *elasticsearch.Client
	ESClientIngredients *elasticsearch.Client
	ESClientTools       *elasticsearch.Client
)

// InitElasticsearch initializes the Elasticsearch clients
func InitElasticsearch() error {
	// Load .env file
	_ = godotenv.Load()

	// Get Elasticsearch configuration from environment variables
	elasticEndpoint := os.Getenv("ELASTIC_SEARCH_ENDPOINT")
	elasticAPIKeyRecipes := os.Getenv("ELASTIC_SEARCH_API_KEY_RECIPES")
	elasticAPIKeyIngredients := os.Getenv("ELASTIC_SEARCH_API_KEY_INGREDIENTS")
	elasticAPIKeyTools := os.Getenv("ELASTIC_SEARCH_API_KEY_TOOLS")

	// Create the Elasticsearch client configuration for ingredients
	cfgIngredients := elasticsearch.Config{
		Addresses: []string{elasticEndpoint},
		APIKey:    elasticAPIKeyIngredients,
	}

	// Create the Elasticsearch client configuration for tools
	cfgTools := elasticsearch.Config{
		Addresses: []string{elasticEndpoint},
		APIKey:    elasticAPIKeyTools,
	}

	// Create the Elasticsearch client configuration for recipes
	cfgRecipes := elasticsearch.Config{
		Addresses: []string{elasticEndpoint},
		APIKey:    elasticAPIKeyRecipes,
	}

	// Create the Elasticsearch clients
	clientIngredients, err := elasticsearch.NewClient(cfgIngredients)
	if err != nil {
		return fmt.Errorf("error creating the Elasticsearch client for ingredients: %w", err)
	}

	// Create the Elasticsearch client for tools
	clientTools, err := elasticsearch.NewClient(cfgTools)
	if err != nil {
		return fmt.Errorf("error creating the Elasticsearch client for tools: %w", err)
	}

	clientRecipes, err := elasticsearch.NewClient(cfgRecipes)
	if err != nil {
		return fmt.Errorf("error creating the Elasticsearch client for recipes: %w", err)
	}

	// Check if the connections are successful
	resIngredients, err := clientIngredients.Info()
	if err != nil {
		return fmt.Errorf("error connecting to Elasticsearch for ingredients: %w", err)
	}
	defer resIngredients.Body.Close()

	// Check if the connections are successful
	resTools, err := clientTools.Info()
	if err != nil {
		return fmt.Errorf("error connecting to Elasticsearch for tools: %w", err)
	}
	defer resTools.Body.Close()

	// Check if the connections are successful
	resRecipes, err := clientRecipes.Info()
	if err != nil {
		return fmt.Errorf("error connecting to Elasticsearch for recipes: %w", err)
	}
	defer resRecipes.Body.Close()

	ESClientIngredients = clientIngredients
	ESClientTools = clientTools
	ESClientRecipes = clientRecipes
	log.Println("Connected to Elasticsearch successfully for both ingredients, tools and recipes!")
	return nil
}

// GetESClientPantries returns the Elasticsearch client for pantries
func GetESClientIngredients() *elasticsearch.Client {
	return ESClientIngredients
}

// GetESClientTools returns the Elasticsearch client for tools
func GetESClientTools() *elasticsearch.Client {
	return ESClientTools
}

// GetESClientRecipes returns the Elasticsearch client for recipes
func GetESClientRecipes() *elasticsearch.Client {
	return ESClientRecipes
}
