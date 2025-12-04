package model

type Item struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Price       int    `json:"price"`
	Description string `json:"description"`
	SoldOut     bool   `json:"sold_out"`
	// ★追加: 画像データを保存・表示するためのフィールド
	ImageURL string `json:"image_url"`
}
