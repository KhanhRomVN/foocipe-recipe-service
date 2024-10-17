package handlers

import (
	"context"

	"github.com/jackc/pgx/v5"
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
