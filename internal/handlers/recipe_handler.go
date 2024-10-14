package handlers

import (
	"net/http"
	"strconv"

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

		c.JSON(http.StatusCreated, gin.H{"message": "Recipe created successfully", "recipe_id": recipeID})
	}
}

func GetRecipe(db *pgxpool.Pool) gin.HandlerFunc {
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

		rows, err := db.Query(c, `
			SELECT pantry_id, quantity
			FROM recipe_ingredient
			WHERE recipe_id = $1
		`, recipeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recipe ingredients"})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var ingredient RecipeIngredientData
			if err := rows.Scan(&ingredient.PantryID, &ingredient.Quantity); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan recipe ingredients"})
				return
			}
			recipe.RecipeIngredientData = append(recipe.RecipeIngredientData, ingredient)
		}

		rows, err = db.Query(c, `
			SELECT pantry_id, quantity
			FROM recipe_tool
			WHERE recipe_id = $1
		`, recipeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recipe tools"})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var tool RecipeToolData
			if err := rows.Scan(&tool.PantryID, &tool.Quantity); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan recipe tools"})
				return
			}
			recipe.RecipeToolData = append(recipe.RecipeToolData, tool)
		}

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

		_, err = tx.Exec(c, `
			UPDATE recipes
			SET name = $1, description = $2, difficulty = $3, prep_time = $4, cook_time = $5,
				servings = $6, category = $7, sub_categories = $8, image_urls = $9, is_public = $10
			WHERE id = $11
		`, req.RecipeData.Name, req.RecipeData.Description, req.RecipeData.Difficulty,
			req.RecipeData.PrepTime, req.RecipeData.CookTime, req.RecipeData.Servings,
			req.RecipeData.Category, req.RecipeData.SubCategories, req.RecipeData.ImageURLs,
			req.RecipeData.IsPublic, recipeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update recipe"})
			return
		}

		_, err = tx.Exec(c, "DELETE FROM recipe_ingredient WHERE recipe_id = $1", recipeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete existing ingredients"})
			return
		}
		_, err = tx.Exec(c, "DELETE FROM recipe_tool WHERE recipe_id = $1", recipeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete existing tools"})
			return
		}
		_, err = tx.Exec(c, "DELETE FROM steps WHERE recipe_id = $1", recipeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete existing steps"})
			return
		}

		if err := insertRecipeIngredients(c, tx, recipeID, req.RecipeIngredientData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert updated ingredients"})
			return
		}
		if err := insertRecipeTools(c, tx, recipeID, req.RecipeToolData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert updated tools"})
			return
		}
		if err := insertSteps(c, tx, recipeID, req.StepsData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert updated steps"})
			return
		}

		if err := tx.Commit(c); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
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

		result, err := tx.Exec(c, "DELETE FROM recipes WHERE id = $1", recipeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete recipe"})
			return
		}

		if result.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
			return
		}

		if err := tx.Commit(c); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Recipe deleted successfully"})
	}
}
