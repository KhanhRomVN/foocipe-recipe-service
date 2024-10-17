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

type Tools struct {
	ID            int      `json:"id"`
	Name          string   `json:"name" binding:"required"`
	Category      string   `json:"category" binding:"required"`
	SubCategories []string `json:"sub_categories"`
	Description   string   `json:"description"`
	Unit          string   `json:"unit"`
	ImageURLs     []string `json:"image_urls"`
}

func CreateTool(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tool Tools
		if err := c.ShouldBindJSON(&tool); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		query := `
            INSERT INTO tools (name, category, sub_categories, description, image_urls)
            VALUES ($1, $2, $3, $4, $5)
            RETURNING id
        `

		var id int
		err := db.QueryRow(c, query,
			tool.Name,
			tool.Category,
			tool.SubCategories,
			tool.Description,
			tool.ImageURLs,
		).Scan(&id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pantry"})
			return
		}

		// Add the pantry to Elasticsearch
		esClient := config.GetESClientIngredients()
		tool.ID = id
		toolJSON, err := json.Marshal(tool)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal pantry data"})
			return
		}

		_, err = esClient.Index(
			"tools",
			bytes.NewReader(toolJSON),
			esClient.Index.WithDocumentID(strconv.Itoa(id)),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to index pantry in Elasticsearch"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Pantry created successfully and indexed in Elasticsearch"})
	}
}

func CreateListTool(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tools []Tools
		if err := c.ShouldBindJSON(&tools); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		createdIDs := make([]int, 0, len(tools))
		esClient := config.GetESClientTools()

		for _, tool := range tools {
			query := `
				INSERT INTO tools (name, category, sub_categories, description, image_urls)
				VALUES ($1, $2, $3, $4, $5)
				RETURNING id
			`

			var id int
			err := db.QueryRow(c, query,
				tool.Name,
				tool.Category,
				tool.SubCategories,
				tool.Description,
				tool.ImageURLs,
			).Scan(&id)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pantry", "details": err.Error()})
				return
			}

			createdIDs = append(createdIDs, id)

			// Add the pantry to Elasticsearch
			tool.ID = id
			toolJSON, err := json.Marshal(tool)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal pantry data", "details": err.Error()})
				return
			}

			_, err = esClient.Index(
				"tools",
				bytes.NewReader(toolJSON),
				esClient.Index.WithDocumentID(strconv.Itoa(id)),
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to index pantry in Elasticsearch", "details": err.Error()})
				return
			}
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Tools created successfully and indexed in Elasticsearch",
			"ids":     createdIDs,
		})
	}
}

func GetToolByID(db *pgxpool.Pool, toolID int) func(*gin.Context) (Tools, error) {
	return func(c *gin.Context) (Tools, error) {
		var tool Tools
		query := `
			SELECT id, name, category, sub_categories, description, image_urls
			FROM tools
			WHERE id = $1
		`

		err := db.QueryRow(c, query, toolID).Scan(
			&tool.ID,
			&tool.Name,
			&tool.Category,
			&tool.SubCategories,
			&tool.Description,
			&tool.ImageURLs,
		)

		if err != nil {
			return Tools{}, err
		}

		return tool, nil
	}
}

func ESSearchTools(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Query("name")

		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Search name is required"})
			return
		}

		esClient := config.GetESClientIngredients()
		searchResult, err := esClient.Search(
			esClient.Search.WithIndex("tools"),
			esClient.Search.WithBody(bytes.NewReader([]byte(`{"query": {"match": {"name": "`+name+`"}}}`))),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search ingredients in Elasticsearch"})
			return
		}

		var result struct {
			Hits struct {
				Hits []struct {
					Source Tools `json:"_source"`
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
