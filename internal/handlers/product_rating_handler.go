package handlers

type ProductRatingData struct {
	ID        int     `json:"id"`
	UserID    int     `json:"user_id"`
	ProductID int     `json:"product_id"`
	Rating    float64 `json:"rating"`
	Comment   string  `json:"comment"`
}

// func GetAverageRating
