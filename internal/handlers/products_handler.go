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

func UpdateProduct(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var product Products
		if err := c.ShouldBindJSON(&product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		id := c.Param("id")
		query := `UPDATE products SET seller_id = $1, title = $2, description = $3, price = $4, stock = $5, image_urls = $6, is_active = $7 WHERE id = $8`
		_, err := db.Exec(c, query, product.SellerID, product.Title, product.Description, product.Price, product.Stock, product.ImageURLs, product.IsActive, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully"})
	}
}

func DeleteProduct(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		query := `DELETE FROM products WHERE id = $1`
		_, err := db.Exec(c, query, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
	}
}

func GetProductByID(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var product Products
		query := `SELECT * FROM products WHERE id = $1`
		err := db.QueryRow(c, query, id).Scan(&product.ID, &product.SellerID, &product.RecipeID, &product.ToolID, &product.IngredientID, &product.Title, &product.Description, &product.Price, &product.Stock, &product.ImageURLs, &product.IsActive)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		c.JSON(http.StatusOK, product)
	}
}

func GetListProduct(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var products []Products
		query := `SELECT * FROM products`
		rows, err := db.Query(c, query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve products"})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var product Products
			if err := rows.Scan(&product.ID, &product.SellerID, &product.RecipeID, &product.ToolID, &product.IngredientID, &product.Title, &product.Description, &product.Price, &product.Stock, &product.ImageURLs, &product.IsActive); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan product"})
				return
			}
			products = append(products, product)
		}

		c.JSON(http.StatusOK, products)
	}
}

func GetProductByRecipeID(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		recipeID := c.Param("id")
		var products []Products
		query := `SELECT * FROM products WHERE recipe_id = $1`
		rows, err := db.Query(c, query, recipeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve products"})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var product Products
			if err := rows.Scan(&product.ID, &product.SellerID, &product.RecipeID, &product.ToolID, &product.IngredientID, &product.Title, &product.Description, &product.Price, &product.Stock, &product.ImageURLs, &product.IsActive); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan product"})
				return
			}
			products = append(products, product)
		}

		c.JSON(http.StatusOK, products)
	}
}

func GetProductByToolID(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		toolID := c.Param("id")
		var products []Products
		query := `SELECT * FROM products WHERE tool_id = $1`
		rows, err := db.Query(c, query, toolID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve products"})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var product Products
			if err := rows.Scan(&product.ID, &product.SellerID, &product.RecipeID, &product.ToolID, &product.IngredientID, &product.Title, &product.Description, &product.Price, &product.Stock, &product.ImageURLs, &product.IsActive); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan product"})
				return
			}
			products = append(products, product)
		}

		c.JSON(http.StatusOK, products)
	}
}

func GetProductByIngredientID(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ingredientID := c.Param("id")
		var products []Products
		query := `SELECT * FROM products WHERE ingredient_id = $1`
		rows, err := db.Query(c, query, ingredientID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve products"})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var product Products
			if err := rows.Scan(&product.ID, &product.SellerID, &product.RecipeID, &product.ToolID, &product.IngredientID, &product.Title, &product.Description, &product.Price, &product.Stock, &product.ImageURLs, &product.IsActive); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan product"})
				return
			}
			products = append(products, product)
		}

		c.JSON(http.StatusOK, products)
	}
}

func GetProductBySellerID(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		sellerID := c.Param("id")
		var products []Products
		query := `SELECT * FROM products WHERE seller_id = $1`
		rows, err := db.Query(c, query, sellerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve products"})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var product Products
			if err := rows.Scan(&product.ID, &product.SellerID, &product.RecipeID, &product.ToolID, &product.IngredientID, &product.Title, &product.Description, &product.Price, &product.Stock, &product.ImageURLs, &product.IsActive); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan product"})
				return
			}
			products = append(products, product)
		}

		c.JSON(http.StatusOK, products)
	}
}
