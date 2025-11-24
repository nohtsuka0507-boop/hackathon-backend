package dao

import (
	"database/sql"
	"hackathon-backend/model" // プロジェクトのモジュール名に合わせて調整が必要な場合があります
	"log"
)

type ItemDAO struct {
	DB *sql.DB
}

func NewItemDAO(db *sql.DB) *ItemDAO {
	return &ItemDAO{DB: db}
}

// 全ての商品を取得する
func (d *ItemDAO) GetAll() ([]*model.Item, error) {
	// sold_out の状態も含めて取得
	rows, err := d.DB.Query("SELECT id, name, price, description, sold_out FROM item")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.Item
	for rows.Next() {
		item := &model.Item{}
		// MySQLのBOOLEANはTINYINT(1)として扱われるため、Scanでboolに変換されます
		if err := rows.Scan(&item.ID, &item.Name, &item.Price, &item.Description, &item.SoldOut); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// 商品を購入する（sold_outフラグをtrueにする）
func (d *ItemDAO) Purchase(id string) error {
	_, err := d.DB.Exec("UPDATE item SET sold_out = TRUE WHERE id = ?", id)
	if err != nil {
		log.Printf("fail: db exec purchase, %v\n", err)
		return err
	}
	return nil
}

// 新しい商品を登録する
func (d *ItemDAO) Insert(item *model.Item) error {
	_, err := d.DB.Exec("INSERT INTO item (id, name, price, description, sold_out) VALUES (?, ?, ?, ?, ?)", item.ID, item.Name, item.Price, item.Description, false)
	if err != nil {
		log.Printf("fail: db exec insert item, %v\n", err)
		return err
	}
	return nil
}
