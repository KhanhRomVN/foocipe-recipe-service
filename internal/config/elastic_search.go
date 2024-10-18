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
	ESClientProducts    *elasticsearch.Client
	ESClientCategories  *elasticsearch.Client
)

// CreateESClient creates an Elasticsearch client with the given API key
func CreateESClient(apiKey string) (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{os.Getenv("ELASTIC_SEARCH_ENDPOINT")},
		APIKey:    apiKey,
	}
	return elasticsearch.NewClient(cfg)
}

// InitElasticsearch initializes the Elasticsearch clients
func InitElasticsearch() error {
	// Load .env file
	_ = godotenv.Load()

	// Get Elasticsearch configuration from environment variables
	elasticAPIKeyRecipes := os.Getenv("ELASTIC_SEARCH_API_KEY_RECIPES")
	elasticAPIKeyIngredients := os.Getenv("ELASTIC_SEARCH_API_KEY_INGREDIENTS")
	elasticAPIKeyTools := os.Getenv("ELASTIC_SEARCH_API_KEY_TOOLS")
	elasticAPIKeyProducts := os.Getenv("ELASTIC_SEARCH_API_KEY_PRODUCTS")
	elasticAPIKeyCategories := os.Getenv("ELASTIC_SEARCH_API_KEY_CATEGORIES")

	// Create the Elasticsearch clients
	var err error
	ESClientIngredients, err = CreateESClient(elasticAPIKeyIngredients)
	if err != nil {
		return fmt.Errorf("error creating the Elasticsearch client for ingredients: %w", err)
	}

	ESClientTools, err = CreateESClient(elasticAPIKeyTools)
	if err != nil {
		return fmt.Errorf("error creating the Elasticsearch client for tools: %w", err)
	}

	ESClientRecipes, err = CreateESClient(elasticAPIKeyRecipes)
	if err != nil {
		return fmt.Errorf("error creating the Elasticsearch client for recipes: %w", err)
	}

	ESClientProducts, err = CreateESClient(elasticAPIKeyProducts)
	if err != nil {
		return fmt.Errorf("error creating the Elasticsearch client for products: %w", err)
	}

	ESClientCategories, err = CreateESClient(elasticAPIKeyCategories)
	if err != nil {
		return fmt.Errorf("error creating the Elasticsearch client for categories: %w", err)
	}

	log.Println("Connected to Elasticsearch successfully for ingredients, tools, recipes, products, and categories!")
	return nil
}

// GetESClientIngredients returns the Elasticsearch client for ingredients
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

// GetESClientProducts returns the Elasticsearch client for products
func GetESClientProducts() *elasticsearch.Client {
	return ESClientProducts
}

// GetESClientCategories returns the Elasticsearch client for categories
func GetESClientCategories() *elasticsearch.Client {
	return ESClientCategories
}
