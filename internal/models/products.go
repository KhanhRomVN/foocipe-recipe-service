package models

type Products struct {
	ID          int      `json:"id"`
	SellerID    int      `json:"seller_id"`
	PantryID    int      `json:"pantry_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	Stock       int      `json:"stock"`
	ImageURLs   []string `json:"image_urls"`
}
