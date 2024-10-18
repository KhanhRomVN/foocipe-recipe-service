package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FavoriteRecipe struct {
	ID       int `json:"id"`
	UserID   int `json:"user_id"`
	RecipeID int `json:"recipe_id"`
}

func CreateFavoriteRecipe(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
