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
	return &ItemDAO{DB: db}
}

// GetAll: 全ての商品を取得 (itemsテーブル)
func (d *ItemDAO) GetAll() ([]*model.Item, error) {
	// ★修正箇所: テーブル名を items (複数形) に指定
	query := "SELECT id, name, price, description, sold_out, image_url FROM items"
	rows, err := d.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.Item
	for rows.Next() {
		item := &model.Item{}
		var imageURL sql.NullString // NULL対策

		// データベースから取得した値を構造体にマッピング
		if err := rows.Scan(&item.ID, &item.Name, &item.Price, &item.Description, &item.SoldOut, &imageURL); err != nil {
			return nil, err
		}

		// 画像URLが有効な場合のみセット
		if imageURL.Valid {
			item.ImageURL = imageURL.String
		}
		items = append(items, item)
	}
	return items, nil
}

// Purchase: 購入処理 (itemsテーブル)
func (d *ItemDAO) Purchase(id string) error {
	// ★修正箇所: テーブル名を items (複数形) に指定
	_, err := d.DB.Exec("UPDATE items SET sold_out = TRUE WHERE id = ?", id)
	if err != nil {
		log.Printf("fail: db exec purchase, %v\n", err)
		return err
	}
	return nil
}

// Insert: 商品登録 (itemsテーブル)
func (d *ItemDAO) Insert(item *model.Item) error {
	// ★修正箇所: テーブル名を items (複数形) に指定
	query := "INSERT INTO items (id, name, price, description, sold_out, image_url) VALUES (?, ?, ?, ?, ?, ?)"

	// 画像URLがない場合は空文字を入れる
	_, err := d.DB.Exec(query, item.ID, item.Name, item.Price, item.Description, false, item.ImageURL)
	if err != nil {
		log.Printf("fail: db exec insert item, %v\n", err)
		return err
	}
	return nil
}
