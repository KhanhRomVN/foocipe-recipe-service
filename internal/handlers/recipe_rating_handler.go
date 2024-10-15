package handlers

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RecipeRatingData struct {
	ID       int     `json:"id"`
	UserID   int     `json:"user_id"`
	RecipeID int     `json:"recipe_id"`
	Rating   float64 `json:"rating"`
	Comment  string  `json:"comment"`
}

func GetAverageRating(db *pgxpool.Pool, recipeID int) (float64, error) {
	var avgRating float64
	err := db.QueryRow(context.Background(), `
		SELECT COALESCE(AVG(rating), 0)
		FROM recipe_rating
		WHERE recipe_id = $1
	`, recipeID).Scan(&avgRating)
	if err != nil {
		return 0, err
	}
	return avgRating, nil
}
