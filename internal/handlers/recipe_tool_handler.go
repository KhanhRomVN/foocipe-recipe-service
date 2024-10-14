package handlers

import (
	"foocipe-recipe-service/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func CreateRecipeTool(tx pgx.Tx, c *gin.Context, tool models.RecipeTool) error {
	query := `INSERT INTO recipe_tools (recipe_id, pantry_id, quantity)
			  VALUES ($1, $2, $3)`

	_, err := tx.Exec(c, query, tool.RecipeID, tool.PantryID, tool.Quantity)
	return err
}

func GetRecipeTools(db *pgx.Conn) gin.HandlerFunc {
	return func(c *gin.Context) {
		recipeID := c.Param("id")

		query := `SELECT id, recipe_id, pantry_id, quantity
				  FROM recipe_tools
				  WHERE recipe_id = $1`

		rows, err := db.Query(c, query, recipeID)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve recipe tools"})
			return
		}
		defer rows.Close()

		var tools []models.RecipeTool
		for rows.Next() {
			var tool models.RecipeTool
			err := rows.Scan(&tool.ID, &tool.RecipeID, &tool.PantryID, &tool.Quantity)
			if err != nil {
				c.JSON(500, gin.H{"error": "Failed to scan recipe tool"})
				return
			}
			tools = append(tools, tool)
		}

		c.JSON(200, tools)
	}
}
