package dao

import (
	"database/sql"
	"hackathon-backend/model"
	"log"
)

type ItemDAO struct {
	DB *sql.DB
}

func NewItemDAO(db *sql.DB) *ItemDAO {
	// ★自動修復機能: アプリ起動時に「like_count」カラムがなければ勝手に追加する
	// (既にカラムがある場合のエラーは無視して進む、というハッカソン用の一時的な処置です)
	_, _ = db.Exec("ALTER TABLE items ADD COLUMN like_count INT DEFAULT 0")

	return &ItemDAO{DB: db}
}

// GetAll: 商品一覧取得（いいね数も含めて取得！）
func (d *ItemDAO) GetAll() ([]*model.Item, error) {
	// クエリに like_count を追加
	query := "SELECT id, name, price, description, sold_out, image_url, like_count FROM items"

	rows, err := d.DB.Query(query)
	if err != nil {
		log.Printf("GetAll Error: %v", err)
		return nil, err
	}
	defer rows.Close()

	return d.scanItems(rows)
}

// Search: 検索機能（いいね数も含めて取得！）
func (d *ItemDAO) Search(keyword string) ([]*model.Item, error) {
	query := "SELECT id, name, price, description, sold_out, image_url, like_count FROM items WHERE name LIKE ?"
	searchTerm := "%" + keyword + "%"

	rows, err := d.DB.Query(query, searchTerm)
	if err != nil {
		log.Printf("Search Error: %v", err)
		return nil, err
	}
	defer rows.Close()

	return d.scanItems(rows)
}

// scanItems: データを読み込む共通処理
func (d *ItemDAO) scanItems(rows *sql.Rows) ([]*model.Item, error) {
	var items []*model.Item
	for rows.Next() {
		item := &model.Item{}
		var imageURL sql.NullString

		// ★修正: データベースから like_count をしっかり読み込む
		if err := rows.Scan(&item.ID, &item.Name, &item.Price, &item.Description, &item.SoldOut, &imageURL, &item.LikeCount); err != nil {
			return nil, err
		}

		if imageURL.Valid {
			item.ImageURL = imageURL.String
		}
		items = append(items, item)
	}
	return items, nil
}

// Purchase: 購入処理
func (d *ItemDAO) Purchase(id string) error {
	_, err := d.DB.Exec("UPDATE items SET sold_out = TRUE WHERE id = ?", id)
	return err
}

// Insert: 商品登録（初期のいいね数は0で登録）
func (d *ItemDAO) Insert(item *model.Item) error {
	// INSERT文にも like_count を追加（デフォルト0）
	query := "INSERT INTO items (id, name, price, description, sold_out, image_url, like_count) VALUES (?, ?, ?, ?, ?, ?, 0)"
	_, err := d.DB.Exec(query, item.ID, item.Name, item.Price, item.Description, false, item.ImageURL)
	return err
}
