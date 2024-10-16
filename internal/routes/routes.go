package routes

import (
	"foocipe-recipe-service/internal/handlers"
	"foocipe-recipe-service/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupRoutes(r *gin.Engine, db *pgxpool.Pool) {
	v1 := r.Group("/v1")

	v1.Use(middleware.AuthToken())
	{
		v1.POST("/recipe", handlers.CreateRecipe(db))
		v1.POST("/pantry", handlers.CreatePantry(db))
		v1.POST("/list-pantry", handlers.CreateListPantry(db))
		v1.GET("/list-recipe", handlers.GetListRecipe(db))
		v1.GET("/recipe/:id", handlers.GetRecipeByID(db))
		v1.GET("/pantries/search", handlers.SearchPantries(db))
	}
}
