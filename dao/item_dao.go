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

// GetAll: シンプルに商品だけを取得（いいね数は一旦0、並び替えなし）
func (d *ItemDAO) GetAll() ([]*model.Item, error) {
	// 一番シンプルなSQLに戻します
	query := "SELECT id, name, price, description, sold_out, image_url FROM items"

	rows, err := d.DB.Query(query)
	if err != nil {
		log.Printf("GetAll Error: %v", err)
		return nil, err
	}
	defer rows.Close()

	return d.scanItems(rows)
}

// Search: 検索機能（こちらもシンプルに）
func (d *ItemDAO) Search(keyword string) ([]*model.Item, error) {
	query := "SELECT id, name, price, description, sold_out, image_url FROM items WHERE name LIKE ?"
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

		// エラーの原因になる「LikeCount」を外して、6項目だけ読み込みます
		if err := rows.Scan(&item.ID, &item.Name, &item.Price, &item.Description, &item.SoldOut, &imageURL); err != nil {
			return nil, err
		}

		if imageURL.Valid {
			item.ImageURL = imageURL.String
		}
		item.LikeCount = 0 // 安全のため0を入れておく
		items = append(items, item)
	}
	return items, nil
}

// Purchase: 購入処理 (変更なし)
func (d *ItemDAO) Purchase(id string) error {
	_, err := d.DB.Exec("UPDATE items SET sold_out = TRUE WHERE id = ?", id)
	return err
}

// Insert: 商品登録 (変更なし)
func (d *ItemDAO) Insert(item *model.Item) error {
	query := "INSERT INTO items (id, name, price, description, sold_out, image_url) VALUES (?, ?, ?, ?, ?, ?)"
	_, err := d.DB.Exec(query, item.ID, item.Name, item.Price, item.Description, false, item.ImageURL)
	return err
}
