package handlers

import (
	"net/http"

	"pleno-recipe-service/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateRecipe(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var recipe models.Recipe
		if err := c.ShouldBindJSON(&recipe); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		recipe.UserID = userID.(int)

		// Here, you would typically insert the recipe into the database
		// For example:
		// err := insertRecipe(db, &recipe)
		// if err != nil {
		//     c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create recipe"})
		//     return
		// }

		c.JSON(http.StatusCreated, recipe)
	}
}

func GetRecipe(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		recipeID := c.Param("id")

		// Here, you would typically fetch the recipe from the database
		// For example:
		// recipe, err := getRecipeByID(db, recipeID, userID.(int))
		// if err != nil {
		//     c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
		//     return
		// }

		// For now, we'll just return a placeholder response
		c.JSON(http.StatusOK, gin.H{
			"message":   "Recipe fetched successfully",
			"recipe_id": recipeID,
			"user_id":   userID,
		})
	}
}
