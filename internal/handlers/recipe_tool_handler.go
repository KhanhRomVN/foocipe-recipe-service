package handlers

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type RecipeToolData struct {
	PantryID   int    `json:"pantry_id"`
	Quantity   string `json:"quantity"`
	PantryName string `json:"pantry_name"`
}

func insertRecipeTools(ctx context.Context, tx pgx.Tx, recipeID int, tools []RecipeToolData) error {
	for _, tool := range tools {
		_, err := tx.Exec(ctx, `
			INSERT INTO recipe_tool (recipe_id, pantry_id, quantity)
			VALUES ($1, $2, $3)
		`, recipeID, tool.PantryID, tool.Quantity)
		if err != nil {
			return err
		}
	}
	return nil
}
