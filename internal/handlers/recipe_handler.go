package handlers

import (
	"net/http"
	"time"

	"foocipe-recipe-service/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateRecipe(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var recipe models.Recipe
		if err := c.ShouldBindJSON(&recipe); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found"})
			return
		}
		recipe.UserID = userID.(int)

		recipe.CreatedAt = time.Now()
		recipe.UpdatedAt = time.Now()

		query := `INSERT INTO recipes (user_id, name, description, difficulty, prep_time, cook_time, servings, category, sub_categories, image_urls, is_public, created_at, updated_at)
				  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
				  RETURNING id`

		err := db.QueryRow(c, query,
			recipe.UserID, recipe.Name, recipe.Description, recipe.Difficulty,
			recipe.PrepTime, recipe.CookTime, recipe.Servings, recipe.Category,
			recipe.SubCategories, recipe.ImageURLs, recipe.IsPublic,
			recipe.CreatedAt, recipe.UpdatedAt).Scan(&recipe.ID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create recipe"})
			return
		}

		c.JSON(http.StatusCreated, recipe)
	}
}

func GetRecipe(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		recipeID := c.Param("id")

		var recipe models.Recipe
		query := `SELECT id, user_id, name, description, difficulty, prep_time, cook_time, servings, category, sub_categories, image_urls, is_public, created_at, updated_at
				  FROM recipes WHERE id = $1`

		err := db.QueryRow(c, query, recipeID).Scan(
			&recipe.ID, &recipe.UserID, &recipe.Name, &recipe.Description,
			&recipe.Difficulty, &recipe.PrepTime, &recipe.CookTime, &recipe.Servings,
			&recipe.Category, &recipe.SubCategories, &recipe.ImageURLs, &recipe.IsPublic,
			&recipe.CreatedAt, &recipe.UpdatedAt)

		if err != nil {
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve recipe"})
			}
			return
		}

		c.JSON(http.StatusOK, recipe)
	}
}
