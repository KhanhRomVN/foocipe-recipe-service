package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RecipeRatingData struct {
	ID       int     `json:"id"`
	UserID   int     `json:"user_id"`
	RecipeID int     `json:"recipe_id"`
	Rating   float64 `json:"rating"`
	Comment  string  `json:"comment"`
}

func CreateRecipeRating(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RecipeRatingData
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := db.Exec(c, `INSERT INTO recipe_rating (user_id, recipe_id, rating, comment) VALUES ($1, $2, $3, $4)`, req.UserID, req.RecipeID, req.Rating, req.Comment)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create recipe rating"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Recipe rating created successfully"})
	}
}

func UpdateRecipeRating(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ratingID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rating ID"})
			return
		}

		var req RecipeRatingData
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err = db.Exec(c, `UPDATE recipe_rating SET rating = $1, comment = $2 WHERE id = $3`, req.Rating, req.Comment, ratingID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update recipe rating"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Recipe rating updated successfully"})
	}
}

func DeleteRecipeRating(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ratingID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rating ID"})
			return
		}

		_, err = db.Exec(c, `DELETE FROM recipe_rating WHERE id = $1`, ratingID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete recipe rating"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Recipe rating deleted successfully"})
	}
}

func GetRecipeRatingByRecipeID(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		recipeID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipe ID"})
			return
		}

		rows, err := db.Query(c, `SELECT * FROM recipe_rating WHERE recipe_id = $1`, recipeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recipe ratings"})
			return
		}
		defer rows.Close()

		var ratings []RecipeRatingData
		for rows.Next() {
			var rating RecipeRatingData
			if err := rows.Scan(&rating.ID, &rating.UserID, &rating.RecipeID, &rating.Rating, &rating.Comment); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan rating data"})
				return
			}
			ratings = append(ratings, rating)
		}

		c.JSON(http.StatusOK, ratings)
	}
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
