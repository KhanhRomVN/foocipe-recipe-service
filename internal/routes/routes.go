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

	setupCartRoutes(v1, db)
	setupCategoriesRoutes(v1, db)
	setupFavoriteRecipeRoutes(v1, db)
	setupIngredientRoutes(v1, db)
	setupProductRatingRoutes(v1, db)
	setupProductRoutes(v1, db)
	setupRecipeRoutes(v1, db)
	setupRecipeIngredientRoutes(v1, db)
	setupRecipeRatingRoutes(v1, db)
	setupRecipeToolRoutes(v1, db)
	setupStepsRoutes(v1, db)
	setupToolRoutes(v1, db)
	setupSearchRoutes(v1, db)
}

func setupCartRoutes(rg *gin.RouterGroup, db *pgxpool.Pool) {
	carts := rg.Group("/carts")
	{
		carts.POST("", handlers.CreateCart(db))
		carts.GET("/:id", handlers.GetCartsByUserID(db))
		carts.PUT("/:id", handlers.UpdateQuantityCart(db))
		carts.DELETE("/:id", handlers.DeleteCartItem(db))
		carts.DELETE("/clear", handlers.DeleteCarts(db))
	}
}

func setupCategoriesRoutes(rg *gin.RouterGroup, db *pgxpool.Pool) {
	categories := rg.Group("/categories")
	{
		categories.POST("", handlers.CreateCategory(db))
		categories.GET("/:id", handlers.GetCategoryByID(db))
		categories.PUT("/:id", handlers.UpdateCategory(db))
		categories.DELETE("/:id", handlers.DeleteCategory(db))
	}
}

func setupRecipeRoutes(rg *gin.RouterGroup, db *pgxpool.Pool) {
	recipes := rg.Group("/recipes")
	{
		recipes.POST("", handlers.CreateRecipe(db))
		recipes.GET("/list", handlers.GetListRecipe(db))
		recipes.GET("/newest", handlers.GetNewestRecipes(db))
		recipes.GET("/my", handlers.GetMyRecipe(db))
		recipes.GET("/:id", handlers.GetRecipeByID(db))
		recipes.PUT("/:id", handlers.UpdateRecipe(db))
		recipes.DELETE("/:id", handlers.DeleteRecipe(db))
		recipes.PUT("/change-owner", handlers.ChangeOwnerRecipe(db))
		recipes.PUT("/change-status", handlers.ChangeStatusRecipe(db))
	}
}

func setupFavoriteRecipeRoutes(rg *gin.RouterGroup, db *pgxpool.Pool) {
	favoriteRecipes := rg.Group("/favorite-recipes")
	{
		favoriteRecipes.POST("", handlers.CreateFavoriteRecipe(db))
	}
}

func setupIngredientRoutes(rg *gin.RouterGroup, db *pgxpool.Pool) {
	ingredients := rg.Group("/ingredients")
	{
		ingredients.POST("", handlers.CreateIngredient(db))
		ingredients.POST("/list", handlers.CreateListIngredient(db))
		ingredients.PUT("/:id", handlers.UpdateIngredient(db))
		// ingredients.DELETE("/:id", handlers.DeleteIngredient(db))
		ingredients.GET("/:id", handlers.GINGetIngredientByID(db))
	}
}

func setupToolRoutes(rg *gin.RouterGroup, db *pgxpool.Pool) {
	tools := rg.Group("/tools")
	{
		tools.POST("", handlers.CreateTool(db))
		tools.POST("/list", handlers.CreateListTool(db))
		tools.PUT("/:id", handlers.UpdateTool(db))
		tools.DELETE("/:id", handlers.DeleteTool(db))
		tools.GET("/:id", handlers.GINGetToolByID(db))
	}
}

func setupProductRoutes(rg *gin.RouterGroup, db *pgxpool.Pool) {
	products := rg.Group("/products")
	{
		products.POST("/create/recipe", handlers.CreateProductAsRecipe(db))
		products.POST("/create/tool", handlers.CreateProductAsTool(db))
		products.POST("/create/ingredient", handlers.CreateProductAsIngredient(db))
		products.PUT("/:id", handlers.UpdateProduct(db))
		products.DELETE("/:id", handlers.DeleteProduct(db))
		products.GET("/:id", handlers.GetProductByID(db))
		products.GET("/list", handlers.GetListProduct(db))
		products.GET("/recipe/:id", handlers.GetProductByRecipeID(db))
		products.GET("/tool/:id", handlers.GetProductByToolID(db))
		products.GET("/ingredient/:id", handlers.GetProductByIngredientID(db))
		products.GET("/seller", handlers.GetProductBySellerID(db))
		products.GET("/newest", handlers.GetNewestProduct(db))
	}
}

func setupRecipeIngredientRoutes(rg *gin.RouterGroup, db *pgxpool.Pool) {
	recipeIngredients := rg.Group("/recipe-ingredients")
	{
		recipeIngredients.PUT("/:id", handlers.UpdateRecipeIngredient(db))
	}
}

func setupRecipeToolRoutes(rg *gin.RouterGroup, db *pgxpool.Pool) {
	recipeTools := rg.Group("/recipe-tools")
	{
		recipeTools.PUT("/:id", handlers.UpdateRecipeTool(db))
	}
}

func setupStepsRoutes(rg *gin.RouterGroup, db *pgxpool.Pool) {
	recipeSteps := rg.Group("/recipe-steps")
	{
		recipeSteps.PUT("/:id", handlers.UpdateRecipeStep(db))
	}
}

func setupRecipeRatingRoutes(rg *gin.RouterGroup, db *pgxpool.Pool) {
	recipeRatings := rg.Group("/recipe-ratings")
	{
		recipeRatings.POST("", handlers.CreateRecipeRating(db))
		recipeRatings.PUT("/:id", handlers.UpdateRecipeRating(db))
		recipeRatings.DELETE("/:id", handlers.DeleteRecipeRating(db))
		recipeRatings.GET("/recipe/:id", handlers.GetRecipeRatingByRecipeID(db))
	}
}

func setupProductRatingRoutes(rg *gin.RouterGroup, db *pgxpool.Pool) {
	productRatings := rg.Group("/product-ratings")
	{
		productRatings.POST("", handlers.CreateProductRating(db))
		productRatings.GET("/reply", handlers.ReplyRating(db))
		productRatings.DELETE("/:id", handlers.DeleteProductRating(db))
		productRatings.PUT("/:id", handlers.UpdateProductRating(db))
		productRatings.GET("/product/:id", handlers.GetProductRatingByProductID(db))
	}
}

func setupSearchRoutes(rg *gin.RouterGroup, db *pgxpool.Pool) {
	search := rg.Group("/search")
	{
		search.GET("/ingredients", handlers.ESSearchIngredients(db))
		search.GET("/tools", handlers.ESSearchTools(db))
		search.GET("/recipes/name", handlers.ESSearchRecipesByName(db))
		search.PUT("/recipes/ingredient", handlers.ESSearchRecipesByIngredient(db))
		// search.GET("/products/name", handlers.ESSearchProductsByName(db))
		// search.GET("/products/recipe", handlers.ESSearchProductsByRecipeID(db))
		// search.GET("/products/tool", handlers.ESSearchProductsByToolID(db))
		// search.GET("/products/ingredient", handlers.ESSearchProductsByIngredientID(db))
		// Only for search in home page (return recipe, ingredient, tool)
		// search.GET("/home", handlers.ESSearchHome(db))
	}
}
