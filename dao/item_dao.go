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
	// ★修正: サブクエリを使って、likesテーブルからその商品のいいね数をカウントします
	query := `
		SELECT 
			i.id, i.name, i.price, i.description, i.sold_out, i.image_url,
			(SELECT COUNT(*) FROM likes WHERE item_id = i.id) as like_count
		FROM items i
	`
	rows, err := d.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.Item
	for rows.Next() {
		item := &model.Item{}
		var imageURL sql.NullString

		// ★修正: 最後の &item.LikeCount を追加して、数を受け取ります
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
	query := "INSERT INTO items (id, name, price, description, sold_out, image_url) VALUES (?, ?, ?, ?, ?, ?)"
	_, err := d.DB.Exec(query, item.ID, item.Name, item.Price, item.Description, false, item.ImageURL)
	if err != nil {
		log.Printf("fail: db exec insert item, %v\n", err)
		return err
	}
	return nil
}
