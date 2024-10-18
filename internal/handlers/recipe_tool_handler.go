package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RecipeToolData struct {
	ToolID   int    `json:"tool_id"`
	Quantity int    `json:"quantity"`
	ToolName string `json:"tool_name,omitempty"`
}

func insertRecipeTools(ctx context.Context, tx pgx.Tx, recipeID int, tools []RecipeToolData) error {
	for _, tool := range tools {
		_, err := tx.Exec(ctx, `
            INSERT INTO recipe_tool (recipe_id, tool_id, quantity)
            VALUES ($1, $2, $3)
        `, recipeID, tool.ToolID, tool.Quantity)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateRecipeTool(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		recipeID, err := strconv.Atoi(c.Param("recipe_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipe ID"})
			return
		}

		var req RecipeToolData
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err = db.Exec(c, `UPDATE recipe_tool SET quantity = $1 WHERE recipe_id = $2 AND tool_id = $3`, req.Quantity, recipeID, req.ToolID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update recipe tool"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Recipe tool updated successfully"})
	}
}
