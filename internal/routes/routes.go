package routes

import (
	"foocipe-recipe-service/internal/handlers"
	"foocipe-recipe-service/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupRoutes(r *gin.Engine, db *pgxpool.Pool) {
	// Tạo nhóm routes v1
	v1 := r.Group("/v1")

	// Áp dụng middleware xác thực cho nhóm v1
	v1.Use(middleware.AuthToken())
	{
		v1.POST("/recipes", handlers.CreateRecipe(db))
		v1.POST("/pantries", handlers.CreatePantries(db))
	}
}
