package model

type Message struct {
	ID        string `json:"id"`
	ItemID    string `json:"item_id"`
	SenderID  string `json:"sender_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}
