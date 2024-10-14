package models

type Categories struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ParentID    int    `json:"parent_id"`
	Level       int    `json:"level"`
}
