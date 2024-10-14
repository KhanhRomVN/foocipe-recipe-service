package models

import "time"

type Recipe struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id" binding:"required"`
	Name          string    `json:"name" binding:"required,min=1,max=255"`
	Description   string    `json:"description" binding:"required"`
	Difficulty    string    `json:"difficulty" binding:"required,oneof=easy medium hard"`
	PrepTime      int       `json:"prep_time" binding:"required,min=0"`
	CookTime      int       `json:"cook_time" binding:"required,min=0"`
	Servings      int       `json:"servings" binding:"required,min=1"`
	Category      string    `json:"category" binding:"required"`
	SubCategories []string  `json:"sub_categories"`
	ImageURLs     []string  `json:"image_urls"`
	IsPublic      bool      `json:"is_public"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
