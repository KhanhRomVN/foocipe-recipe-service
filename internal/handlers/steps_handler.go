package handlers

import (
	"context"

	"github.com/jackc/pgx/v5"
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
