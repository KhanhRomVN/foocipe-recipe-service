package handlers

import (
	"foocipe-recipe-service/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreatePantries(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var pantry models.Pantries

		// Bind JSON request body to the pantry struct
		if err := c.ShouldBindJSON(&pantry); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// SQL query to insert the new pantry
		query := `
			INSERT INTO pantries (name, category, sub_categories, description, image_urls)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id`

		// Execute the query
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

		// Set the ID of the newly created pantry
		pantry.ID = id

		// Return the created pantry with status 201 (Created)
		c.JSON(http.StatusCreated, pantry)
	}
}
