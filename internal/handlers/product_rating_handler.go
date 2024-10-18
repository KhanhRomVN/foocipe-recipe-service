package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRatingData struct {
	ID        int     `json:"id"`
	UserID    int     `json:"user_id"`
	ProductID int     `json:"product_id"`
	Rating    float64 `json:"rating"`
	Comment   string  `json:"comment"`
	ReplyID   int     `json:"reply_id"`
}

func CreateProductRating(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ratingData ProductRatingData
		if err := c.ShouldBindJSON(&ratingData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Logic to save ratingData to the database
		c.JSON(http.StatusCreated, ratingData)
	}
}

func ReplyRating(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Reply added"})
	}
}

func DeleteProductRating(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNoContent, nil)
	}
}

func UpdateProductRating(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ratingData ProductRatingData
		if err := c.ShouldBindJSON(&ratingData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, ratingData)
	}
}

func GetProductRatingByProductID(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ratings": []ProductRatingData{}})
	}
}
