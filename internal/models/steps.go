package models

type Steps struct {
	ID          int    `json:"id"`
	RecipeID    int    `json:"recipe_id"`
	StepNumber  int    `json:"step_number"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
