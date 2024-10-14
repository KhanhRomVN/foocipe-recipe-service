package models

type RecipeTool struct {
	ID       int    `json:"id"`
	RecipeID int    `json:"recipe_id"`
	ToolID   int    `json:"tool_id"`
	Quantity string `json:"quantity"`
}
