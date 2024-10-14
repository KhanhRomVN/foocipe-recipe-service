package routes

import (
	"foocipe-recipe-service/internal/handlers"
	"foocipe-recipe-service/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupRoutes(r *gin.Engine, db *pgxpool.Pool) {
	protected := r.Group("/")
	protected.Use(middleware.AuthToken())
	{
		protected.POST("/recipes", handlers.CreateRecipe(db))
		protected.GET("/recipes/:id", handlers.GetRecipe(db))
	}
}
