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
	// 1. まず、既にいいねしているか確認
	var exists bool
	queryCheck := "SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = ? AND item_id = ?)"
	err := dao.db.QueryRow(queryCheck, userID, itemID).Scan(&exists)
	if err != nil {
		return false, err
	}

	if exists {
		// --- 既にいいね済みなら -> 解除 (Unlike) ---

		// A. likesテーブルから削除
		_, err := dao.db.Exec("DELETE FROM likes WHERE user_id = ? AND item_id = ?", userID, itemID)
		if err != nil {
			return false, err
		}

		// B. itemsテーブルのいいね数(like_count)を -1 する
		// ★ここが「数が保存されない」を直すための重要な行です！
		_, err = dao.db.Exec("UPDATE items SET like_count = like_count - 1 WHERE id = ?", itemID)
		if err != nil {
			return false, err
		}

		return false, nil // "いいね解除"

	} else {
		// --- まだいいねしてないなら -> 登録 (Like) ---

		// A. likesテーブルに登録
		_, err := dao.db.Exec("INSERT INTO likes (user_id, item_id) VALUES (?, ?)", userID, itemID)
		if err != nil {
			return false, err
		}

		// B. itemsテーブルのいいね数(like_count)を +1 する
		// ★ここが「数が保存されない」を直すための重要な行です！
		_, err = dao.db.Exec("UPDATE items SET like_count = like_count + 1 WHERE id = ?", itemID)
		if err != nil {
			return false, err
		}

		return true, nil // "いいね"
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
