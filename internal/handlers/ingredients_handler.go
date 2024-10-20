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

type Ingredients struct {
	ID            int      `json:"id"`
	Name          string   `json:"name" binding:"required"`
	Category      string   `json:"category" binding:"required"`
	SubCategories []string `json:"sub_categories"`
	Description   string   `json:"description" binding:"required"`
	ImageURLs     []string `json:"image_urls"`
	Unit          string   `json:"unit" binding:"required"`
}

func CreateIngredient(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ingredient Ingredients
		if err := c.ShouldBindJSON(&ingredient); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Add the ingredient to PostgreSQL
		query := `
			INSERT INTO ingredients (name, category, sub_categories, description, image_urls, unit)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`
		var id int
		err := db.QueryRow(c, query,
			ingredient.Name,
			ingredient.Category,
			ingredient.SubCategories,
			ingredient.Description,
			ingredient.ImageURLs,
			ingredient.Unit,
		).Scan(&id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ingredient", "details": err.Error()})
			return
		}

		// Add the ingredient to Elasticsearch
		esClient := config.GetESClientIngredients()
		ingredient.ID = id
		ingredientJSON, err := json.Marshal(ingredient)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal ingredient data"})
			return
		}

		_, err = esClient.Index(
			"ingredients",
			bytes.NewReader(ingredientJSON),
			esClient.Index.WithDocumentID(strconv.Itoa(id)),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to index ingredient in Elasticsearch"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Ingredient created successfully and indexed in Elasticsearch"})
	}
}

func CreateListIngredient(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ingredients []Ingredients
		if err := c.ShouldBindJSON(&ingredients); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		createdIDs := make([]int, 0, len(ingredients))
		esClient := config.GetESClientIngredients()

		for _, ingredient := range ingredients {
			query := `
				INSERT INTO ingredients (name, category, sub_categories, description, image_urls, unit)
				VALUES ($1, $2, $3, $4, $5, $6)
				RETURNING id
			`

			var id int
			err := db.QueryRow(c, query,
				ingredient.Name,
				ingredient.Category,
				ingredient.SubCategories,
				ingredient.Description,
				ingredient.ImageURLs,
				ingredient.Unit,
			).Scan(&id)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ingredient", "details": err.Error()})
				return
			}

			createdIDs = append(createdIDs, id)

			// Add the ingredient to Elasticsearch
			ingredient.ID = id
			ingredientJSON, err := json.Marshal(ingredient)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal ingredient data", "details": err.Error()})
				return
			}

			_, err = esClient.Index(
				"ingredients",
				bytes.NewReader(ingredientJSON),
				esClient.Index.WithDocumentID(strconv.Itoa(id)),
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to index ingredient in Elasticsearch", "details": err.Error()})
				return
			}
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Ingredients created successfully and indexed in Elasticsearch",
			"ids":     createdIDs,
		})
	}
}

func UpdateIngredient(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ingredientID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ingredient ID"})
			return
		}

		var req Ingredients
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		query := `
			UPDATE ingredients SET
			name = $1, category = $2, sub_categories = $3, description = $4, image_urls = $5, unit = $6
			WHERE id = $7
		`
		_, err = db.Exec(c, query, req.Name, req.Category, req.SubCategories, req.Description, req.ImageURLs, req.Unit, ingredientID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update ingredient"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Ingredient updated successfully"})
	}
}

func GetIngredientByID(db *pgxpool.Pool, ingredientID int) func(*gin.Context) (Ingredients, error) {
	return func(c *gin.Context) (Ingredients, error) {
		var ingredient Ingredients
		query := `
			SELECT id, name, category, sub_categories, description, image_urls, unit
			FROM ingredients
			WHERE id = $1
		`

		err := db.QueryRow(c, query, ingredientID).Scan(
			&ingredient.ID,
			&ingredient.Name,
			&ingredient.Category,
			&ingredient.SubCategories,
			&ingredient.Description,
			&ingredient.ImageURLs,
			&ingredient.Unit,
		)

		if err != nil {
			return Ingredients{}, err
		}

		return ingredient, nil
	}
}

func GINGetIngredientByID(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ingredientID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ingredient ID"})
			return
		}

		var ingredient Ingredients
		query := `SELECT id, name, category, sub_categories, description, image_urls, unit FROM ingredients WHERE id = $1`
		err = db.QueryRow(c, query, ingredientID).Scan(&ingredient.ID, &ingredient.Name, &ingredient.Category, &ingredient.SubCategories, &ingredient.Description, &ingredient.ImageURLs, &ingredient.Unit)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Ingredient not found"})
			return
		}

		c.JSON(http.StatusOK, ingredient)
	}
}

func ESSearchIngredients(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Query("name")

		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Search name is required"})
			return
		}

		esClient := config.GetESClientIngredients()
		searchResult, err := esClient.Search(
			esClient.Search.WithIndex("ingredients"),
			esClient.Search.WithBody(bytes.NewReader([]byte(`{"query": {"match": {"name": "`+name+`"}}}`))),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search ingredients in Elasticsearch"})
			return
		}

		var result struct {
			Hits struct {
				Hits []struct {
					Source Ingredients `json:"_source"`
				} `json:"hits"`
			} `json:"hits"`
		}

		if err := json.NewDecoder(searchResult.Body).Decode(&result); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode search results"})
			return
		}

		c.JSON(http.StatusOK, result.Hits.Hits)
	}
}
