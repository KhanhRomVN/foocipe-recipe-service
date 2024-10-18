package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StepData struct {
	StepNumber  int    `json:"step_number"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func insertSteps(ctx context.Context, tx pgx.Tx, recipeID int, steps []StepData) error {
	for _, step := range steps {
		_, err := tx.Exec(ctx, `
			INSERT INTO steps (recipe_id, step_number, title, description)
			VALUES ($1, $2, $3, $4)
		`, recipeID, step.StepNumber, step.Title, step.Description)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateRecipeStep(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		recipeID, err := strconv.Atoi(c.Param("recipe_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipe ID"})
			return
		}

		var req StepData
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err = db.Exec(c, `UPDATE steps SET title = $1, description = $2 WHERE recipe_id = $3 AND step_number = $4`, req.Title, req.Description, recipeID, req.StepNumber)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update recipe step"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Recipe step updated successfully"})
	}
}
