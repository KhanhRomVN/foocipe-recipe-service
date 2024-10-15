package handlers

import (
	"context"
	"net/http"

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

		c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Pantry created successfully"})
	}
}

func GetPantryByID(db *pgxpool.Pool, id int) (Pantry, error) {
	var pantry Pantry
	err := db.QueryRow(context.Background(), `
		SELECT id, name, category, sub_categories, description, image_urls
		FROM pantries
		WHERE id = $1
	`, id).Scan(&pantry.ID, &pantry.Name, &pantry.Category, &pantry.SubCategories, &pantry.Description, &pantry.ImageURLs)
	return pantry, err
}
