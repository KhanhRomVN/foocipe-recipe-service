package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

func UpdateRecipeIngredient(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		recipeID, err := strconv.Atoi(c.Param("recipe_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipe ID"})
			return
		}

		var req RecipeIngredientData
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err = db.Exec(c, `UPDATE recipe_ingredient SET quantity = $1 WHERE recipe_id = $2 AND ingredient_id = $3`, req.Quantity, recipeID, req.IngredientID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update recipe ingredient"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Recipe ingredient updated successfully"})
	}
}
