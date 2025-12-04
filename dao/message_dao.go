package dao

import (
	"database/sql"
	"fmt"
	"hackathon-backend/model"
)

type MessageDAO struct {
	db *sql.DB
}

func NewMessageDAO(db *sql.DB) *MessageDAO {
	return &MessageDAO{db: db}
}

// 取得: created_at の昇順（古い順）で取得
func (dao *MessageDAO) GetByItemID(itemID string) ([]*model.Message, error) {
	query := "SELECT id, item_id, sender_id, content, created_at FROM messages WHERE item_id = ? ORDER BY created_at ASC"

	rows, err := dao.db.Query(query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*model.Message
	for rows.Next() {
		var m model.Message
		if err := rows.Scan(&m.ID, &m.ItemID, &m.SenderID, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, &m)
	}
	return messages, nil
}

// 保存
func (dao *MessageDAO) Insert(msg *model.Message) error {
	query := "INSERT INTO messages (id, item_id, sender_id, content, created_at) VALUES (?, ?, ?, ?, ?)"
	_, err := dao.db.Exec(query, msg.ID, msg.ItemID, msg.SenderID, msg.Content, msg.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert message: %w", err)
	}
	return nil
}
