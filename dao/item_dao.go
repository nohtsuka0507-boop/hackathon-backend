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

// GetAll: 全ての商品と「いいね数」を取得
func (d *ItemDAO) GetAll() ([]*model.Item, error) {
	query := `
		SELECT 
			i.id, i.name, i.price, i.description, i.sold_out, i.image_url,
			(SELECT COUNT(*) FROM likes WHERE item_id = i.id) as like_count
		FROM items i
		ORDER BY i.created_at DESC
	` // ※新しい順(DESC)に並ぶように修正しました（created_atがない場合は i.id DESC でもOK）
	// もし created_at カラムがないエラーが出る場合は ORDER BY i.id DESC にしてください。

	rows, err := d.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return d.scanItems(rows)
}

// ★追加: 商品名で検索 (部分一致)
func (d *ItemDAO) Search(keyword string) ([]*model.Item, error) {
	// キーワードを含む商品を検索 (いいね数も一緒に取得)
	query := `
		SELECT 
			i.id, i.name, i.price, i.description, i.sold_out, i.image_url,
			(SELECT COUNT(*) FROM likes WHERE item_id = i.id) as like_count
		FROM items i
		WHERE i.name LIKE ?
		ORDER BY i.id DESC
	`
	searchTerm := "%" + keyword + "%"

	rows, err := d.DB.Query(query, searchTerm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return d.scanItems(rows)
}

// 共通のスキャン処理
func (d *ItemDAO) scanItems(rows *sql.Rows) ([]*model.Item, error) {
	var items []*model.Item
	for rows.Next() {
		item := &model.Item{}
		var imageURL sql.NullString

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
	if err != nil {
		log.Printf("fail: db exec purchase, %v\n", err)
		return err
	}
	return nil
}

// Insert: 商品登録
func (d *ItemDAO) Insert(item *model.Item) error {
	// created_at があれば追加すべきですが、今回はシンプルな構成を維持します
	query := "INSERT INTO items (id, name, price, description, sold_out, image_url) VALUES (?, ?, ?, ?, ?, ?)"
	_, err := d.DB.Exec(query, item.ID, item.Name, item.Price, item.Description, false, item.ImageURL)
	if err != nil {
		log.Printf("fail: db exec insert item, %v\n", err)
		return err
	}
	return nil
}
