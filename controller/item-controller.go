package controller

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"hackathon-backend/dao"
	"hackathon-backend/model"
	"log"
	"net/http"
)

type ItemController struct {
	ItemDAO *dao.ItemDAO
}

func NewItemController(itemDAO *dao.ItemDAO) *ItemController {
	return &ItemController{ItemDAO: itemDAO}
}

// HandleGetItems: 商品一覧または検索結果を返す (GET /items)
func (c *ItemController) HandleGetItems(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// ★追加: URLクエリパラメータ "q" を取得
	keyword := r.URL.Query().Get("q")

	var items []*model.Item
	var err error

	if keyword != "" {
		// キーワードがある場合は検索を実行
		log.Printf("Searching items with keyword: %s", keyword)
		items, err = c.ItemDAO.Search(keyword)
	} else {
		// ない場合は全件取得
		items, err = c.ItemDAO.GetAll()
	}

	if err != nil {
		log.Printf("fail: get items, %v\n", err)
		// エラー時も空配列を返してフロントエンドがクラッシュしないようにする
		json.NewEncoder(w).Encode([]*model.Item{})
		return
	}

	// nullの場合は空配列にする
	if items == nil {
		items = []*model.Item{}
	}

	log.Printf("Success: Returning %d items\n", len(items))
	json.NewEncoder(w).Encode(items)
}

// HandlePurchase: 商品を購入する
func (c *ItemController) HandlePurchase(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := c.ItemDAO.Purchase(id); err != nil {
		log.Printf("fail: purchase item, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Purchase successful"}`)
}

// HandleAddItem: 商品を出品する
func (c *ItemController) HandleAddItem(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling AddItem request...")

	var req struct {
		Name        string `json:"name"`
		Price       int    `json:"price"`
		Description string `json:"description"`
		ImageURL    string `json:"image_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("fail: decode request body, %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, _ := generateItemID()

	item := &model.Item{
		ID:          id,
		Name:        req.Name,
		Price:       req.Price,
		Description: req.Description,
		ImageURL:    req.ImageURL,
	}

	if err := c.ItemDAO.Insert(item); err != nil {
		log.Printf("fail: insert item, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Println("Success: Item inserted into DB")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func generateItemID() (string, error) {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}
