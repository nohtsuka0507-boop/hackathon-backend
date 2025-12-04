package controller

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"hackathon-backend/dao"
	"hackathon-backend/model"
	"net/http"
	"time"
)

type ChatController struct {
	MessageDAO *dao.MessageDAO // ポインタとして扱う
}

// コンストラクタ
func NewChatController(messageDAO *dao.MessageDAO) *ChatController {
	return &ChatController{MessageDAO: messageDAO}
}

// HandleGetMessages: メッセージ取得 (GET /messages?item_id=xxx)
func (c *ChatController) HandleGetMessages(w http.ResponseWriter, r *http.Request) {
	itemID := r.URL.Query().Get("item_id")
	if itemID == "" {
		http.Error(w, "item_id is required", http.StatusBadRequest)
		return
	}

	msgs, err := c.MessageDAO.GetByItemID(itemID)
	if err != nil {
		// エラーログはサーバー側に出すとデバッグしやすい
		fmt.Printf("DB Error: %v\n", err)
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}

	if msgs == nil {
		msgs = []*model.Message{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msgs)
}

// HandlePostMessage: メッセージ送信 (POST /messages)
func (c *ChatController) HandlePostMessage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ItemID   string `json:"item_id"`
		SenderID string `json:"sender_id"`
		Content  string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	id, _ := generateID()

	msg := &model.Message{
		ID:       id,
		ItemID:   req.ItemID,
		SenderID: req.SenderID,
		Content:  req.Content,
		// ★修正: 日付も含めたフォーマットに変更（並び順のため重要！）
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	if err := c.MessageDAO.Insert(msg); err != nil {
		fmt.Printf("Save Error: %v\n", err)
		http.Error(w, "Save Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msg)
}

func generateID() (string, error) {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}
