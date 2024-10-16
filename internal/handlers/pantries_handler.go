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

type Pantry struct {
	ID            int      `json:"id"`
	Name          string   `json:"name" binding:"required"`
	Category      string   `json:"category" binding:"required"`
	SubCategories []string `json:"sub_categories"`
	Description   string   `json:"description"`
	ImageURLs     []string `json:"image_urls"`
}

func CreatePantry(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var pantry Pantry
		if err := c.ShouldBindJSON(&pantry); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		query := `
			INSERT INTO pantries (name, category, sub_categories, description, image_urls)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`

		var id int
		err := db.QueryRow(c, query,
			pantry.Name,
			pantry.Category,
			pantry.SubCategories,
			pantry.Description,
			pantry.ImageURLs,
		).Scan(&id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pantry"})
			return
		}

		// Add the pantry to Elasticsearch
		esClient := config.GetESClientPantries()
		pantry.ID = id
		pantryJSON, err := json.Marshal(pantry)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal pantry data"})
			return
		}

		_, err = esClient.Index(
			"pantries",
			bytes.NewReader(pantryJSON),
			esClient.Index.WithDocumentID(strconv.Itoa(id)),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to index pantry in Elasticsearch"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Pantry created successfully and indexed in Elasticsearch"})
	}
}

func CreateListPantry(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var pantries []Pantry
		if err := c.ShouldBindJSON(&pantries); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		createdIDs := make([]int, 0, len(pantries))
		esClient := config.GetESClientPantries()

		for _, pantry := range pantries {
			query := `
				INSERT INTO pantries (name, category, sub_categories, description, image_urls)
				VALUES ($1, $2, $3, $4, $5)
				RETURNING id
			`

			var id int
			err := db.QueryRow(c, query,
				pantry.Name,
				pantry.Category,
				pantry.SubCategories,
				pantry.Description,
				pantry.ImageURLs,
			).Scan(&id)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pantry", "details": err.Error()})
				return
			}

			createdIDs = append(createdIDs, id)

			// Add the pantry to Elasticsearch
			pantry.ID = id
			pantryJSON, err := json.Marshal(pantry)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal pantry data", "details": err.Error()})
				return
			}

			_, err = esClient.Index(
				"pantries",
				bytes.NewReader(pantryJSON),
				esClient.Index.WithDocumentID(strconv.Itoa(id)),
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to index pantry in Elasticsearch", "details": err.Error()})
				return
			}
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Pantries created successfully and indexed in Elasticsearch",
			"ids":     createdIDs,
		})
	}
}

func GetPantryByID(db *pgxpool.Pool, pantryID int) func(*gin.Context) (Pantry, error) {
	return func(c *gin.Context) (Pantry, error) {
		var pantry Pantry
		query := `
			SELECT id, name, category, sub_categories, description, image_urls
			FROM pantries
			WHERE id = $1
		`

		err := db.QueryRow(c, query, pantryID).Scan(
			&pantry.ID,
			&pantry.Name,
			&pantry.Category,
			&pantry.SubCategories,
			&pantry.Description,
			&pantry.ImageURLs,
		)

		if err != nil {
			return Pantry{}, err
		}

		return pantry, nil
	}
}

func ESSearchPantries(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Query("name")

		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Search name is required"})
			return
		}

		esClient := config.GetESClientPantries()
		searchResult, err := esClient.Search(
			esClient.Search.WithIndex("pantries"),
			esClient.Search.WithBody(bytes.NewReader([]byte(`{"query": {"match": {"name": "`+name+`"}}}`))),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search pantries in Elasticsearch"})
			return
		}

		var result struct {
			Hits struct {
				Hits []struct {
					Source Pantry `json:"_source"`
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
