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
	Description   string   `json:"description"`
	Unit          string   `json:"unit"`
	ImageURLs     []string `json:"image_urls"`
}

func CreateIngredient(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ingredient Ingredients
		if err := c.ShouldBindJSON(&ingredient); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		query := `
			INSERT INTO ingredients (name, category, sub_categories, description, image_urls)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`

		var id int
		err := db.QueryRow(c, query,
			ingredient.Name,
			ingredient.Category,
			ingredient.SubCategories,
			ingredient.Description,
			ingredient.ImageURLs,
		).Scan(&id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pantry"})
			return
		}

		// Add the pantry to Elasticsearch
		esClient := config.GetESClientIngredients()
		ingredient.ID = id
		ingredientJSON, err := json.Marshal(ingredient)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal pantry data"})
			return
		}

		_, err = esClient.Index(
			"ingredients",
			bytes.NewReader(ingredientJSON),
			esClient.Index.WithDocumentID(strconv.Itoa(id)),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to index pantry in Elasticsearch"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Pantry created successfully and indexed in Elasticsearch"})
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
				INSERT INTO ingredients (name, category, sub_categories, description, image_urls)
				VALUES ($1, $2, $3, $4, $5)
				RETURNING id
			`

			var id int
			err := db.QueryRow(c, query,
				ingredient.Name,
				ingredient.Category,
				ingredient.SubCategories,
				ingredient.Description,
				ingredient.ImageURLs,
			).Scan(&id)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pantry", "details": err.Error()})
				return
			}

			createdIDs = append(createdIDs, id)

			// Add the pantry to Elasticsearch
			ingredient.ID = id
			ingredientJSON, err := json.Marshal(ingredient)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal pantry data", "details": err.Error()})
				return
			}

			_, err = esClient.Index(
				"ingredients",
				bytes.NewReader(ingredientJSON),
				esClient.Index.WithDocumentID(strconv.Itoa(id)),
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to index pantry in Elasticsearch", "details": err.Error()})
				return
			}
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Ingredients created successfully and indexed in Elasticsearch",
			"ids":     createdIDs,
		})
	}
}

func GetIngredientByID(db *pgxpool.Pool, ingredientID int) func(*gin.Context) (Ingredients, error) {
	return func(c *gin.Context) (Ingredients, error) {
		var ingredient Ingredients
		query := `
			SELECT id, name, category, sub_categories, description, image_urls
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
		)

		if err != nil {
			return Ingredients{}, err
		}

		return ingredient, nil
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
