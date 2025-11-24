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

// 商品一覧を返す (GET /items)
func (c *ItemController) HandleGetItems(w http.ResponseWriter, r *http.Request) {
	items, err := c.ItemDAO.GetAll()
	if err != nil {
		log.Printf("fail: get all items, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// 商品を購入する (POST /items/purchase)
func (c *ItemController) HandlePurchase(w http.ResponseWriter, r *http.Request) {
	// URLクエリからIDを取得する場合: /items/purchase?id=1
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

// 商品を出品する (POST /items)
func (c *ItemController) HandleAddItem(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Price       int    `json:"price"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ID生成 (簡易的なランダムID)
	id, _ := generateItemID()

	item := &model.Item{
		ID:          id,
		Name:        req.Name,
		Price:       req.Price,
		Description: req.Description,
	}

	if err := c.ItemDAO.Insert(item); err != nil {
		log.Printf("fail: insert item, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// 簡易的なID生成関数
func generateItemID() (string, error) {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}
