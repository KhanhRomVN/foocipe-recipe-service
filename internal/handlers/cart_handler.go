package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Cart struct {
	ID        int `json:"id"`
	UserID    int `json:"user_id"`
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

func CreateCart(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ProductID int `json:"product_id" binding:"required"`
			Quantity  int `json:"quantity" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
}

func GetCartsByUserID(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		rows, err := db.Query(c, `SELECT * FROM carts WHERE user_id = $1`, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch carts"})
			return
		}
		defer rows.Close()

		var carts []Cart
		for rows.Next() {
			var cart Cart
			if err := rows.Scan(&cart.ID, &cart.UserID, &cart.ProductID, &cart.Quantity); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan cart data"})
				return
			}
			carts = append(carts, cart)
		}

		c.JSON(http.StatusOK, carts)
	}
}

func UpdateQuantityCart(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		cartID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cart ID"})
			return
		}

		var req struct {
			Quantity int `json:"quantity" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err = db.Exec(c, `UPDATE carts SET quantity = $1 WHERE id = $2`, req.Quantity, cartID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart quantity"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Cart quantity updated successfully"})
	}
}

func DeleteCartItem(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		cartID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cart ID"})
			return
		}

		_, err = db.Exec(c, `DELETE FROM carts WHERE id = $1`, cartID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete cart item"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Cart item deleted successfully"})
	}
}

func DeleteCarts(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		_, err := db.Exec(c, `DELETE FROM carts WHERE user_id = $1`, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete carts"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "All carts deleted successfully"})
	}
}
