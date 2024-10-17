package handlers

import (
	"bytes"
	"encoding/json"
	"foocipe-recipe-service/internal/config"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Products struct {
	ID           int      `json:"id"`
	SellerID     int      `json:"seller_id"`
	RecipeID     int      `json:"recipe_id"`
	ToolID       int      `json:"tool_id"`
	IngredientID int      `json:"ingredient_id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Price        int      `json:"price"`
	Stock        int      `json:"stock"`
	ImageURLs    []string `json:"image_urls"`
	IsActive     bool     `json:"is_active"`
}

func CreateProductAsRecipe(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var product Products
		if err := c.ShouldBindJSON(&product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Ensure RecipeID is provided
		if product.RecipeID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "RecipeID is required"})
			return
		}

		// Insert the product into the database
		query := `INSERT INTO products (seller_id, recipe_id, title, description, price, stock, image_urls, is_active)
                  VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

		var id int
		err := db.QueryRow(c, query, product.SellerID, product.RecipeID, product.Title, product.Description,
			product.Price, product.Stock, product.ImageURLs, product.IsActive).Scan(&id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
			return
		}

		// Index the product in Elasticsearch
		if err := indexProductInElasticsearch(c, id, product); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to index product in Elasticsearch"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Product created successfully"})
	}
}

func CreateProductAsTool(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var product Products
		if err := c.ShouldBindJSON(&product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Ensure ToolID is provided
		if product.ToolID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ToolID is required"})
			return
		}

		// Insert the product into the database
		query := `INSERT INTO products (seller_id, tool_id, title, description, price, stock, image_urls, is_active)
                  VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

		var id int
		err := db.QueryRow(c, query, product.SellerID, product.ToolID, product.Title, product.Description,
			product.Price, product.Stock, product.ImageURLs, product.IsActive).Scan(&id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
			return
		}

		// Index the product in Elasticsearch
		if err := indexProductInElasticsearch(c, id, product); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to index product in Elasticsearch"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Product created successfully"})
	}
}

func CreateProductAsIngredient(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var product Products
		if err := c.ShouldBindJSON(&product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Ensure IngredientID is provided
		if product.IngredientID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "IngredientID is required"})
			return
		}

		// Insert the product into the database
		query := `INSERT INTO products (seller_id, ingredient_id, title, description, price, stock, image_urls, is_active)
                  VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

		var id int
		err := db.QueryRow(c, query, product.SellerID, product.IngredientID, product.Title, product.Description,
			product.Price, product.Stock, product.ImageURLs, product.IsActive).Scan(&id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
			return
		}

		// Index the product in Elasticsearch
		if err := indexProductInElasticsearch(c, id, product); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to index product in Elasticsearch"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Product created successfully"})
	}
}

func indexProductInElasticsearch(c *gin.Context, id int, product Products) error {
	esClient := config.GetESClientProducts()
	product.ID = id
	productJSON, err := json.Marshal(product)
	if err != nil {
		return err
	}

	_, err = esClient.Index(
		"products",
		bytes.NewReader(productJSON),
		esClient.Index.WithDocumentID(strconv.Itoa(id)),
	)
	return err
}
