package models

type Pantries struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	Category      string   `json:"category"`
	SubCategories []string `json:"sub_categories"`
	Description   string   `json:"description"`
	ImageURLs     []string `json:"image_urls"`
}
