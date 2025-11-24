package model

type Item struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Price       int    `json:"price"`
	Description string `json:"description"`
	SoldOut     bool   `json:"sold_out"`
}
