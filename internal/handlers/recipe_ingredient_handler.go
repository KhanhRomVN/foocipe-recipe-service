package handlers

import (
	"foocipe-recipe-service/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func CreateRecipeIngredient(tx pgx.Tx, c *gin.Context, ingredient models.RecipeIngredient) error {
	query := `INSERT INTO recipe_ingredients (recipe_id, pantry_id, quantity)
              VALUES ($1, $2, $3)`

	_, err := tx.Exec(c, query, ingredient.RecipeID, ingredient.PantryID, ingredient.Quantity)
	return err
}

func GetRecipeIngredients(db *pgx.Conn) gin.HandlerFunc {
	return func(c *gin.Context) {
		recipeID := c.Param("id")

		query := `SELECT id, recipe_id, pantry_id, quantity
				  FROM recipe_ingredients
				  WHERE recipe_id = $1`

		rows, err := db.Query(c, query, recipeID)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve recipe ingredients"})
			return
		}
		defer rows.Close()

		var ingredients []models.RecipeIngredient
		for rows.Next() {
			var ingredient models.RecipeIngredient
			err := rows.Scan(&ingredient.ID, &ingredient.RecipeID, &ingredient.PantryID, &ingredient.Quantity)
			if err != nil {
				c.JSON(500, gin.H{"error": "Failed to scan recipe ingredient"})
				return
			}
			ingredients = append(ingredients, ingredient)
		}

		c.JSON(200, ingredients)
	}
}
