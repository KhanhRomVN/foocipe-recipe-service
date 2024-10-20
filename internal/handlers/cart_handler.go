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

func AddProductToCart(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		productID, err := strconv.Atoi(c.Param("product_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}

		quantity, err := strconv.Atoi(c.Param("quantity"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quantity"})
			return
		}

		// Thêm sản phẩm vào giỏ hàng
		_, err = db.Exec(c, `INSERT INTO carts (user_id, product_id, quantity) VALUES ($1, $2, $3)`, userID, productID, quantity)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add product to cart"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Product added to cart successfully"})
	}
}

type CartResponse struct {
	ID           int      `json:"id"`
	UserID       int      `json:"user_id"`
	ProductID    int      `json:"product_id"`
	Quantity     int      `json:"quantity"`
	IngredientID *int     `json:"ingredient_id"`
	ToolID       *int     `json:"tool_id"`
	RecipeID     *int     `json:"recipe_id"`
	Title        string   `json:"title"`
	Price        float64  `json:"price"`
	Stock        int      `json:"stock"`
	ImageURLs    []string `json:"image_urls"`
}

func GetCartsByUserID(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Update the query to join with the products table
		query := `
			SELECT c.id, c.user_id, c.product_id, c.quantity, 
			       p.ingredient_id, p.tool_id, p.recipe_id, p.title, p.price, p.stock, p.image_urls
			FROM carts c
			JOIN products p ON c.product_id = p.id
			WHERE c.user_id = $1`

		rows, err := db.Query(c, query, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch carts"})
			return
		}
		defer rows.Close()

		var carts []CartResponse

		for rows.Next() {
			var cart CartResponse
			if err := rows.Scan(&cart.ID, &cart.UserID, &cart.ProductID, &cart.Quantity,
				&cart.IngredientID, &cart.ToolID, &cart.RecipeID, &cart.Title,
				&cart.Price, &cart.Stock, &cart.ImageURLs); err != nil {
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
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		cartID, err := strconv.Atoi(c.Param("cart_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cart ID"})
			return
		}

		quantity, err := strconv.Atoi(c.Param("quantity"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quantity"})
			return
		}

		_, err = db.Exec(c, `UPDATE carts SET quantity = $1 WHERE id = $2 AND user_id = $3`, quantity, cartID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart quantity"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Cart quantity updated successfully"})
	}
}

func DeleteCartItem(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		cartID, err := strconv.Atoi(c.Param("cart_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cart ID"})
			return
		}

		_, err = db.Exec(c, `DELETE FROM carts WHERE id = $1 AND user_id = $2`, cartID, userID)
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear carts"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "All cart items cleared successfully"})
	}
}
