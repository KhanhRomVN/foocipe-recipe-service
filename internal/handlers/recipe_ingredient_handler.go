package handlers

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type RecipeIngredientData struct {
	PantryID int `json:"pantry_id"`
	Quantity int `json:"quantity"`
}

func insertRecipeIngredients(ctx context.Context, tx pgx.Tx, recipeID int, ingredients []RecipeIngredientData) error {
	for _, ingredient := range ingredients {
		_, err := tx.Exec(ctx, `
			INSERT INTO recipe_ingredient (recipe_id, pantry_id, quantity)
			VALUES ($1, $2, $3)
		`, recipeID, ingredient.PantryID, ingredient.Quantity)
		if err != nil {
			return err
		}
	}
	return nil
}
