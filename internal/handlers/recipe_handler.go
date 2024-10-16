package handlers

import (
	"bytes"
	"encoding/json"
	"foocipe-recipe-service/internal/config"
	"net/http"
	"strconv"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RecipeRequest struct {
	RecipeData struct {
		Name          string   `json:"name"`
		Description   string   `json:"description"`
		Difficulty    string   `json:"difficulty"`
		PrepTime      int      `json:"prep_time"`
		CookTime      int      `json:"cook_time"`
		Servings      int      `json:"servings"`
		Category      string   `json:"category"`
		SubCategories []string `json:"sub_categories"`
		ImageURLs     []string `json:"image_urls"`
		IsPublic      bool     `json:"is_public"`
	} `json:"recipeData"`
	RecipeIngredientData []RecipeIngredientData `json:"recipeIngredientData"`
	RecipeToolData       []RecipeToolData       `json:"recipeToolData"`
	StepsData            []StepData             `json:"stepsData"`
}

func CreateRecipe(db *pgxpool.Pool) gin.HandlerFunc {
	esClient := config.GetESClientRecipes()
	return func(c *gin.Context) {
		var req RecipeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		tx, err := db.Begin(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}
		defer tx.Rollback(c)

		var recipeID int

		err = tx.QueryRow(c, `
			INSERT INTO recipes (user_id, name, description, difficulty, prep_time, cook_time, servings, category, sub_categories, image_urls, is_public)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			RETURNING id
		`, userID, req.RecipeData.Name, req.RecipeData.Description, req.RecipeData.Difficulty,
			req.RecipeData.PrepTime, req.RecipeData.CookTime, req.RecipeData.Servings, req.RecipeData.Category,
			req.RecipeData.SubCategories, req.RecipeData.ImageURLs, req.RecipeData.IsPublic).Scan(&recipeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert recipe"})
			return
		}

		if err := insertRecipeIngredients(c, tx, recipeID, req.RecipeIngredientData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert recipe ingredients"})
			return
		}

		if err := insertRecipeTools(c, tx, recipeID, req.RecipeToolData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert recipe tools"})
			return
		}

		if err := insertSteps(c, tx, recipeID, req.StepsData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert recipe steps"})
			return
		}

		if err := tx.Commit(c); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		// Index the recipe in Elasticsearch
		if err := indexRecipeInElasticsearch(c, db, esClient, recipeID, req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to index recipe in Elasticsearch"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Recipe created successfully and indexed in Elasticsearch", "recipe_id": recipeID})
	}
}

func indexRecipeInElasticsearch(c *gin.Context, db *pgxpool.Pool, esClient *elasticsearch.Client, recipeID int, req RecipeRequest) error {
	// Prepare the recipe data for Elasticsearch
	esRecipe := map[string]interface{}{
		"id":             recipeID,
		"name":           req.RecipeData.Name,
		"description":    req.RecipeData.Description,
		"difficulty":     req.RecipeData.Difficulty,
		"prep_time":      req.RecipeData.PrepTime,
		"cook_time":      req.RecipeData.CookTime,
		"servings":       req.RecipeData.Servings,
		"category":       req.RecipeData.Category,
		"sub_categories": req.RecipeData.SubCategories,
		"image_urls":     req.RecipeData.ImageURLs,
		"is_public":      req.RecipeData.IsPublic,
	}

	// Process recipe ingredients
	ingredients := make([]map[string]interface{}, 0)
	for _, ing := range req.RecipeIngredientData {
		pantry, err := GetPantryByID(db, ing.PantryID)(c)
		if err != nil {
			return err
		}
		ingredients = append(ingredients, map[string]interface{}{
			"pantry_id":   ing.PantryID,
			"quantity":    ing.Quantity,
			"pantry_name": pantry.Name,
		})
	}
	esRecipe["ingredients"] = ingredients

	// Process recipe tools
	tools := make([]map[string]interface{}, 0)
	for _, tool := range req.RecipeToolData {
		pantry, err := GetPantryByID(db, tool.PantryID)(c)
		if err != nil {
			return err
		}
		tools = append(tools, map[string]interface{}{
			"pantry_id":   tool.PantryID,
			"quantity":    tool.Quantity,
			"pantry_name": pantry.Name,
		})
	}
	esRecipe["tools"] = tools

	// Add steps data
	esRecipe["steps"] = req.StepsData

	// Convert the recipe data to JSON
	recipeJSON, err := json.Marshal(esRecipe)
	if err != nil {
		return err
	}

	// Index the recipe in Elasticsearch
	_, err = esClient.Index(
		"recipes",
		bytes.NewReader(recipeJSON),
		esClient.Index.WithDocumentID(strconv.Itoa(recipeID)),
	)
	if err != nil {
		return err
	}

	return nil
}

func GetListRecipe(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(c, `
			SELECT id, name, difficulty, cook_time, image_urls
			FROM recipes
			ORDER BY id DESC
			LIMIT 10
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recipes"})
			return
		}
		defer rows.Close()

		var recipes []gin.H
		for rows.Next() {
			var id int
			var name, difficulty string
			var cookTime int
			var imageURLs []string
			if err := rows.Scan(&id, &name, &difficulty, &cookTime, &imageURLs); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan recipe data"})
				return
			}

			// Get average rating
			avgRating, err := GetAverageRating(db, id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get average rating"})
				return
			}

			recipes = append(recipes, gin.H{
				"id":             id,
				"name":           name,
				"difficulty":     difficulty,
				"cook_time":      cookTime,
				"image_urls":     imageURLs,
				"average_rating": avgRating,
			})
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating over recipes"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"recipes": recipes})
	}
}

func GetRecipeByID(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		recipeID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipe ID"})
			return
		}

		var recipe RecipeRequest
		err = db.QueryRow(c, `
			SELECT name, description, difficulty, prep_time, cook_time, servings, category, sub_categories, image_urls, is_public
			FROM recipes
			WHERE id = $1
		`, recipeID).Scan(
			&recipe.RecipeData.Name, &recipe.RecipeData.Description, &recipe.RecipeData.Difficulty,
			&recipe.RecipeData.PrepTime, &recipe.RecipeData.CookTime, &recipe.RecipeData.Servings,
			&recipe.RecipeData.Category, &recipe.RecipeData.SubCategories, &recipe.RecipeData.ImageURLs,
			&recipe.RecipeData.IsPublic,
		)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
			return
		}

		// Fetch recipe ingredients
		rows, err := db.Query(c, `
			SELECT ri.pantry_id, ri.quantity, p.name
			FROM recipe_ingredient ri
			JOIN pantries p ON ri.pantry_id = p.id
			WHERE ri.recipe_id = $1
		`, recipeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recipe ingredients"})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var ingredient RecipeIngredientData
			if err := rows.Scan(&ingredient.PantryID, &ingredient.Quantity, &ingredient.PantryName); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan recipe ingredients"})
				return
			}
			recipe.RecipeIngredientData = append(recipe.RecipeIngredientData, ingredient)
		}

		// Fetch recipe tools
		rows, err = db.Query(c, `
			SELECT rt.pantry_id, rt.quantity, p.name
			FROM recipe_tool rt
			JOIN pantries p ON rt.pantry_id = p.id
			WHERE rt.recipe_id = $1
		`, recipeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recipe tools"})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var tool RecipeToolData
			if err := rows.Scan(&tool.PantryID, &tool.Quantity, &tool.PantryName); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan recipe tools"})
				return
			}
			recipe.RecipeToolData = append(recipe.RecipeToolData, tool)
		}

		// Fetch recipe steps
		rows, err = db.Query(c, `
			SELECT step_number, title, description
			FROM steps
			WHERE recipe_id = $1
			ORDER BY step_number
		`, recipeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recipe steps"})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var step StepData
			if err := rows.Scan(&step.StepNumber, &step.Title, &step.Description); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan recipe steps"})
				return
			}
			recipe.StepsData = append(recipe.StepsData, step)
		}

		c.JSON(http.StatusOK, recipe)
	}
}

func ESSearchRecipesByName(db *pgxpool.Pool) gin.HandlerFunc {
	esClient := config.GetESClientRecipes()
	return func(c *gin.Context) {
		query := c.Query("name")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
			return
		}

		// Prepare the search request
		var buf bytes.Buffer
		searchQuery := map[string]interface{}{
			"query": map[string]interface{}{
				"match": map[string]interface{}{
					"name": query,
				},
			},
		}
		if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode search query"})
			return
		}

		// Perform the search request
		res, err := esClient.Search(
			esClient.Search.WithContext(c.Request.Context()),
			esClient.Search.WithIndex("recipes"),
			esClient.Search.WithBody(&buf),
			esClient.Search.WithTrackTotalHits(true),
			esClient.Search.WithPretty(),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute search"})
			return
		}
		defer res.Body.Close()

		if res.IsError() {
			var e map[string]interface{}
			if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse the response body"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": e})
			return
		}

		var r map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse the response body"})
			return
		}

		// Extract and format the search results
		hits := r["hits"].(map[string]interface{})["hits"].([]interface{})
		recipes := make([]gin.H, len(hits))

		for i, hit := range hits {
			source := hit.(map[string]interface{})["_source"].(map[string]interface{})
			recipes[i] = gin.H{
				"id":             source["id"],
				"name":           source["name"],
				"description":    source["description"],
				"difficulty":     source["difficulty"],
				"prep_time":      source["prep_time"],
				"cook_time":      source["cook_time"],
				"servings":       source["servings"],
				"category":       source["category"],
				"sub_categories": source["sub_categories"],
				"image_urls":     source["image_urls"],
				"is_public":      source["is_public"],
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"total":   r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"],
			"recipes": recipes,
		})
	}
}
