package models

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type RecipeIngredient struct {
	ID       int `json:"id"`
	RecipeID int `json:"recipe_id" binding:"required"`
	PantryID int `json:"pantry_id" binding:"required"`
	Quantity int `json:"quantity" binding:"required,min=1"`
}

func insertRecipeIngredients(ctx context.Context, tx pgx.Tx, recipeID int, ingredients []RecipeIngredient) error {
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
