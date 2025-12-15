package controller

import (
	"encoding/json"
	"hackathon-backend/dao"
	"log"
	"net/http"
)

type LikeController struct {
	LikeDAO *dao.LikeDAO
}

func NewLikeController(likeDAO *dao.LikeDAO) *LikeController {
	return &LikeController{LikeDAO: likeDAO}
}

// HandleToggleLike: いいねの切り替え
func (c *LikeController) HandleToggleLike(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
		ItemID string `json:"item_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	isLiked, err := c.LikeDAO.ToggleLike(req.UserID, req.ItemID)
	if err != nil {
		log.Printf("【いいねエラー】ToggleLike Failed: %v", err)
		http.Error(w, "Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"liked": isLiked})
}

// HandleGetLikes: いいね一覧の取得
func (c *LikeController) HandleGetLikes(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}

	ids, err := c.LikeDAO.GetLikedItemIDs(userID)
	if err != nil {
		log.Printf("【いいね取得エラー】GetLikedItemIDs Failed: %v", err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	if ids == nil {
		ids = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ids)
}
