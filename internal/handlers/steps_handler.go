package handlers

import (
	"foocipe-recipe-service/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func CreateStep(tx pgx.Tx, c *gin.Context, step models.Steps) error {
	query := `INSERT INTO steps (recipe_id, step_number, title, description)
			  VALUES ($1, $2, $3, $4)`

	_, err := tx.Exec(c, query, step.RecipeID, step.StepNumber, step.Title, step.Description)
	return err
}

func GetRecipeSteps(db *pgx.Conn) gin.HandlerFunc {
	return func(c *gin.Context) {
		recipeID := c.Param("id")

		query := `SELECT id, recipe_id, step_number, title, description
				  FROM steps
				  WHERE recipe_id = $1
				  ORDER BY step_number`

		rows, err := db.Query(c, query, recipeID)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to retrieve recipe steps"})
			return
		}
		defer rows.Close()

		var steps []models.Steps
		for rows.Next() {
			var step models.Steps
			err := rows.Scan(&step.ID, &step.RecipeID, &step.StepNumber, &step.Title, &step.Description)
			if err != nil {
				c.JSON(500, gin.H{"error": "Failed to scan recipe step"})
				return
			}
			steps = append(steps, step)
		}

		c.JSON(200, steps)
	}
}
