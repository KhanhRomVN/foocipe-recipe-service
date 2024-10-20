package handlers

import (
	"bytes"
	"encoding/json"
	"foocipe-recipe-service/internal/config"
	"net/http"
	"strconv"
	"strings"

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

		// insert data to table recipe_ingredients
		if err := insertRecipeIngredients(c, tx, recipeID, req.RecipeIngredientData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert recipe ingredients"})
			return
		}

		// insert data to table recipe_tools
		if err := insertRecipeTools(c, tx, recipeID, req.RecipeToolData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert recipe tools"})
			return
		}

		// insert data to table steps
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
		ingredient, err := GetIngredientByID(db, ing.IngredientID)(c)
		if err != nil {
			return err
		}
		ingredients = append(ingredients, map[string]interface{}{
			"ingredient_id":   ing.IngredientID,
			"quantity":        ing.Quantity,
			"ingredient_name": ingredient.Name,
		})
	}
	esRecipe["ingredients"] = ingredients

	// Process recipe tools
	tools := make([]map[string]interface{}, 0)
	for _, tool := range req.RecipeToolData {
		toolData, err := GetToolByID(db, tool.ToolID)(c)
		if err != nil {
			return err
		}
		tools = append(tools, map[string]interface{}{
			"tool_id":   tool.ToolID,
			"quantity":  tool.Quantity,
			"tool_name": toolData.Name,
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

func GetNewestRecipes(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(c, `
			SELECT id, name, difficulty, cook_time, image_urls
			FROM recipes
			ORDER BY id DESC
			LIMIT 10
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch newest recipes"})
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

		c.JSON(http.StatusOK, recipes)
	}
}

func GetMyRecipe(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		rows, err := db.Query(c, `
			SELECT id, name, difficulty, cook_time, image_urls
			FROM recipes
			WHERE user_id = $1
			ORDER BY id DESC
		`, userID)
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

			recipes = append(recipes, gin.H{
				"id":         id,
				"name":       name,
				"difficulty": difficulty,
				"cook_time":  cookTime,
				"image_urls": imageURLs,
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

		// Fetch Recipe
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
			SELECT ri.ingredient_id, ri.quantity, i.name, i.unit
			FROM recipe_ingredient ri
			JOIN ingredients i ON ri.ingredient_id = i.id
			WHERE ri.recipe_id = $1
		`, recipeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recipe ingredients"})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var ingredient RecipeIngredientData
			if err := rows.Scan(&ingredient.IngredientID, &ingredient.Quantity, &ingredient.IngredientName, &ingredient.Unit); err != nil { // Thêm &ingredient.Unit
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan recipe ingredients"})
				return
			}
			recipe.RecipeIngredientData = append(recipe.RecipeIngredientData, ingredient)
		}

		// Fetch recipe tools
		rows, err = db.Query(c, `
			SELECT rt.tool_id, rt.quantity, t.name, t.unit
			FROM recipe_tool rt
			JOIN tools t ON rt.tool_id = t.id
			WHERE rt.recipe_id = $1
		`, recipeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recipe tools"})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var tool RecipeToolData
			if err := rows.Scan(&tool.ToolID, &tool.Quantity, &tool.ToolName, &tool.Unit); err != nil { // Thêm &tool.Unit
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

func UpdateRecipe(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		recipeID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipe ID"})
			return
		}

		var req RecipeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		tx, err := db.Begin(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}
		defer tx.Rollback(c)

		// Update recipe
		_, err = tx.Exec(c, `
            UPDATE recipes SET
            name = $1, description = $2, difficulty = $3, prep_time = $4,
            cook_time = $5, servings = $6, category = $7, sub_categories = $8,
            image_urls = $9, is_public = $10
            WHERE id = $11
        `, req.RecipeData.Name, req.RecipeData.Description, req.RecipeData.Difficulty,
			req.RecipeData.PrepTime, req.RecipeData.CookTime, req.RecipeData.Servings,
			req.RecipeData.Category, req.RecipeData.SubCategories, req.RecipeData.ImageURLs,
			req.RecipeData.IsPublic, recipeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update recipe"})
			return
		}

		if err := tx.Commit(c); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		// Update Elasticsearch index
		if err := indexRecipeInElasticsearch(c, db, config.GetESClientRecipes(), recipeID, req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update recipe in Elasticsearch"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Recipe updated successfully"})
	}
}

func DeleteRecipe(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		recipeID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipe ID"})
			return
		}

		tx, err := db.Begin(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}
		defer tx.Rollback(c)

		// Delete related data
		_, err = tx.Exec(c, "DELETE FROM recipe_ingredient WHERE recipe_id = $1", recipeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete recipe ingredients"})
			return
		}
		_, err = tx.Exec(c, "DELETE FROM recipe_tool WHERE recipe_id = $1", recipeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete recipe tools"})
			return
		}
		_, err = tx.Exec(c, "DELETE FROM steps WHERE recipe_id = $1", recipeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete recipe steps"})
			return
		}

		// Delete the recipe
		_, err = tx.Exec(c, "DELETE FROM recipes WHERE id = $1", recipeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete recipe"})
			return
		}

		if err := tx.Commit(c); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		// Remove from Elasticsearch
		esClient := config.GetESClientRecipes()
		_, err = esClient.Delete("recipes", strconv.Itoa(recipeID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete recipe from Elasticsearch"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Recipe deleted successfully"})
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

func ChangeOwnerRecipe(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RecipeID   int `json:"recipe_id" binding:"required"`
			NewOwnerID int `json:"new_owner_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := db.Exec(c, "UPDATE recipes SET user_id = $1 WHERE id = $2", req.NewOwnerID, req.RecipeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change recipe owner"})
			return
		}

		// Update Elasticsearch
		esClient := config.GetESClientRecipes()
		_, err = esClient.Update("recipes", strconv.Itoa(req.RecipeID),
			strings.NewReader(`{"doc": {"user_id": `+strconv.Itoa(req.NewOwnerID)+`}}`))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update recipe owner in Elasticsearch"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Recipe owner changed successfully"})
	}
}

func ChangeStatusRecipe(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RecipeID int  `json:"recipe_id" binding:"required"`
			IsPublic bool `json:"is_public" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := db.Exec(c, "UPDATE recipes SET is_public = $1 WHERE id = $2", req.IsPublic, req.RecipeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change recipe status"})
			return
		}

		// Update Elasticsearch
		esClient := config.GetESClientRecipes()
		_, err = esClient.Update("recipes", strconv.Itoa(req.RecipeID),
			strings.NewReader(`{"doc": {"is_public": `+strconv.FormatBool(req.IsPublic)+`}}`))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update recipe status in Elasticsearch"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Recipe status changed successfully"})
	}
}

func ESSearchRecipesByIngredient(db *pgxpool.Pool) gin.HandlerFunc {
	esClient := config.GetESClientRecipes()
	return func(c *gin.Context) {
		var reqBody struct {
			Ingredients []int `json:"ingredients" binding:"required"`
		}

		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if len(reqBody.Ingredients) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "At least one ingredient ID is required"})
			return
		}

		// Convert []int to []interface{} for Elasticsearch query
		ingredientIDs := make([]interface{}, len(reqBody.Ingredients))
		for i, id := range reqBody.Ingredients {
			ingredientIDs[i] = id
		}

		// Prepare the search request
		var buf bytes.Buffer
		searchQuery := map[string]interface{}{
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"must": []map[string]interface{}{
						{
							"terms": map[string]interface{}{
								"ingredients.ingredient_id": ingredientIDs,
							},
						},
					},
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
				"ingredients":    source["ingredients"],
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"total":   r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"],
			"recipes": recipes,
		})
	}
}

// func UpdateRecipe(db *pgxpool.Pool) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		recipeID, err := strconv.Atoi(c.Param("id"))
// 		if err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipe ID"})
// 			return
// 		}

// 		var req RecipeRequest
// 		if err := c.ShouldBindJSON(&req); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}

// 		tx, err := db.Begin(c)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
// 			return
// 		}
// 		defer tx.Rollback(c)

// 		// Update recipe
// 		_, err = tx.Exec(c, `
//             UPDATE recipes SET
//             name = $1, description = $2, difficulty = $3, prep_time = $4,
//             cook_time = $5, servings = $6, category = $7, sub_categories = $8,
//             image_urls = $9, is_public = $10
//             WHERE id = $11
//         `, req.RecipeData.Name, req.RecipeData.Description, req.RecipeData.Difficulty,
// 			req.RecipeData.PrepTime, req.RecipeData.CookTime, req.RecipeData.Servings,
// 			req.RecipeData.Category, req.RecipeData.SubCategories, req.RecipeData.ImageURLs,
// 			req.RecipeData.IsPublic, recipeID)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update recipe"})
// 			return
// 		}

// 		// Update ingredients, tools, and steps
// 		if err := updateRecipeIngredients(c, tx, recipeID, req.RecipeIngredientData); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update recipe ingredients"})
// 			return
// 		}
// 		if err := updateRecipeTools(c, tx, recipeID, req.RecipeToolData); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update recipe tools"})
// 			return
// 		}
// 		if err := updateSteps(c, tx, recipeID, req.StepsData); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update recipe steps"})
// 			return
// 		}

// 		if err := tx.Commit(c); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
// 			return
// 		}

// 		// Update Elasticsearch index
// 		if err := indexRecipeInElasticsearch(c, db, config.GetESClientRecipes(), recipeID, req); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update recipe in Elasticsearch"})
// 			return
// 		}

// 		c.JSON(http.StatusOK, gin.H{"message": "Recipe updated successfully"})
// 	}
// }

// func DeleteRecipe(db *pgxpool.Pool) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		recipeID, err := strconv.Atoi(c.Param("id"))
// 		if err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipe ID"})
// 			return
// 		}

// 		tx, err := db.Begin(c)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
// 			return
// 		}
// 		defer tx.Rollback(c)

// 		// Delete related data
// 		_, err = tx.Exec(c, "DELETE FROM recipe_ingredient WHERE recipe_id = $1", recipeID)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete recipe ingredients"})
// 			return
// 		}
// 		_, err = tx.Exec(c, "DELETE FROM recipe_tool WHERE recipe_id = $1", recipeID)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete recipe tools"})
// 			return
// 		}
// 		_, err = tx.Exec(c, "DELETE FROM steps WHERE recipe_id = $1", recipeID)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete recipe steps"})
// 			return
// 		}

// 		// Delete the recipe
// 		_, err = tx.Exec(c, "DELETE FROM recipes WHERE id = $1", recipeID)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete recipe"})
// 			return
// 		}

// 		if err := tx.Commit(c); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
// 			return
// 		}

// 		// Remove from Elasticsearch
// 		esClient := config.GetESClientRecipes()
// 		_, err = esClient.Delete("recipes", strconv.Itoa(recipeID))
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete recipe from Elasticsearch"})
// 			return
// 		}

// 		c.JSON(http.StatusOK, gin.H{"message": "Recipe deleted successfully"})
// 	}
// }

// func ChangeOwnerRecipe(db *pgxpool.Pool) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		var req struct {
// 			RecipeID   int `json:"recipe_id" binding:"required"`
// 			NewOwnerID int `json:"new_owner_id" binding:"required"`
// 		}
// 		if err := c.ShouldBindJSON(&req); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}

// 		_, err := db.Exec(c, "UPDATE recipes SET user_id = $1 WHERE id = $2", req.NewOwnerID, req.RecipeID)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change recipe owner"})
// 			return
// 		}

// 		// Update Elasticsearch
// 		esClient := config.GetESClientRecipes()
// 		_, err = esClient.Update("recipes", strconv.Itoa(req.RecipeID),
// 			strings.NewReader(`{"doc": {"user_id": `+strconv.Itoa(req.NewOwnerID)+`}}`))
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update recipe owner in Elasticsearch"})
// 			return
// 		}

// 		c.JSON(http.StatusOK, gin.H{"message": "Recipe owner changed successfully"})
// 	}
// }

// func ChangeStatusRecipe(db *pgxpool.Pool) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		var req struct {
// 			RecipeID int  `json:"recipe_id" binding:"required"`
// 			IsPublic bool `json:"is_public" binding:"required"`
// 		}
// 		if err := c.ShouldBindJSON(&req); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}

// 		_, err := db.Exec(c, "UPDATE recipes SET is_public = $1 WHERE id = $2", req.IsPublic, req.RecipeID)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change recipe status"})
// 			return
// 		}

// 		// Update Elasticsearch
// 		esClient := config.GetESClientRecipes()
// 		_, err = esClient.Update("recipes", strconv.Itoa(req.RecipeID),
// 			strings.NewReader(`{"doc": {"is_public": `+strconv.FormatBool(req.IsPublic)+`}}`))
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update recipe status in Elasticsearch"})
// 			return
// 		}

// 		c.JSON(http.StatusOK, gin.H{"message": "Recipe status changed successfully"})
// 	}
// }
