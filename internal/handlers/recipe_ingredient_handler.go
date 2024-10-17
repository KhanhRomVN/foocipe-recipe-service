package handlers

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type RecipeIngredientData struct {
	IngredientID   int    `json:"ingredient_id"`
	Quantity       int    `json:"quantity"`
	IngredientName string `json:"ingredient_name,omitempty"`
}

func insertRecipeIngredients(ctx context.Context, tx pgx.Tx, recipeID int, ingredients []RecipeIngredientData) error {
	for _, ingredient := range ingredients {
		_, err := tx.Exec(ctx, `
            INSERT INTO recipe_ingredient (recipe_id, ingredient_id, quantity)
            VALUES ($1, $2, $3)
        `, recipeID, ingredient.IngredientID, ingredient.Quantity)
		if err != nil {
			return err
		}
	}
	return nil
}
