package models

type RecipeTool struct {
	ID       int    `json:"id"`
	RecipeID int    `json:"recipe_id"`
	PantryID int    `json:"pantry_id"`
	Quantity string `json:"quantity"`
}
