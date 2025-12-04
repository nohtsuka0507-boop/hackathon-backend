package model

type Item struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Price       int    `json:"price"`
	Description string `json:"description"`
	SoldOut     bool   `json:"sold_out"`
	ImageURL    string `json:"image_url"`
	// ★追加: いいねの数を格納するフィールド
	LikeCount int `json:"like_count"`
}
