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
		v1.POST("/ingredient", handlers.CreateIngredient(db))
		v1.POST("/list-ingredient", handlers.CreateListIngredient(db))
		v1.POST("/tool", handlers.CreateTool(db))
		v1.POST("/list-tool", handlers.CreateListTool(db))
		v1.GET("/list-recipe", handlers.GetListRecipe(db))
		v1.GET("/recipe/:id", handlers.GetRecipeByID(db))
		v1.GET("/ingredients/search", handlers.ESSearchIngredients(db))
		v1.GET("/tools/search", handlers.ESSearchTools(db))
		v1.GET("/ingredients/search", handlers.ESSearchIngredients(db))
		v1.GET("/tools/search", handlers.ESSearchTools(db))
		v1.GET("/recipes/name/search", handlers.ESSearchRecipesByName(db))
		v1.PUT("/recipes/ingredient/search", handlers.ESSearchRecipesByIngredient(db))
	}
}
