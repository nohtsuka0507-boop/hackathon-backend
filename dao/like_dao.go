package dao

import (
	"database/sql"
)

type LikeDAO struct {
	db *sql.DB
}

func NewLikeDAO(db *sql.DB) *LikeDAO {
	return &LikeDAO{db: db}
}

// ToggleLike: いいねを付けたり消したりする（スイッチ機能）
func (dao *LikeDAO) ToggleLike(userID, itemID string) (bool, error) {
	// まず、既にいいねしているか確認
	var exists bool
	queryCheck := "SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = ? AND item_id = ?)"
	err := dao.db.QueryRow(queryCheck, userID, itemID).Scan(&exists)
	if err != nil {
		return false, err
	}

	if exists {
		// 既にいいね済みなら -> 削除 (Unlike)
		_, err := dao.db.Exec("DELETE FROM likes WHERE user_id = ? AND item_id = ?", userID, itemID)
		if err != nil {
			return false, err
		}
		return false, nil // "いいね解除" したので false を返す
	} else {
		// まだいいねしてないなら -> 登録 (Like)
		_, err := dao.db.Exec("INSERT INTO likes (user_id, item_id) VALUES (?, ?)", userID, itemID)
		if err != nil {
			return false, err
		}
		return true, nil // "いいね" したので true を返す
	}
}

// GetLikedItemIDs: そのユーザーがいいねした商品のID一覧を取得
func (dao *LikeDAO) GetLikedItemIDs(userID string) ([]string, error) {
	rows, err := dao.db.Query("SELECT item_id FROM likes WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
