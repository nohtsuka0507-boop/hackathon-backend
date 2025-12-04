package controller

import (
	"encoding/json"
	"fmt"
	"hackathon-backend/dao"
	"hackathon-backend/model"
	"net/http"
)

type AuthController struct {
	UserDAO *dao.UserDAO
}

// ★修正: sql.DB ではなく UserDAO を受け取るように変更
func NewAuthController(userDAO *dao.UserDAO) *AuthController {
	return &AuthController{UserDAO: userDAO}
}

// HandleRegister: ユーザー登録
func (c *AuthController) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := model.NewUser(req.Name, req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.UserDAO.CreateUser(user); err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// HandleLogin: ログイン
func (c *AuthController) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 1. Emailでユーザーを検索
	user, err := c.UserDAO.GetUserByEmail(req.Email)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not found or invalid password", http.StatusUnauthorized)
		return
	}

	// 2. パスワードの一致確認
	if user.Password != req.Password {
		http.Error(w, "User not found or invalid password", http.StatusUnauthorized)
		return
	}

	// ログイン成功
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Login successful", "user_id": "%s", "name": "%s"}`, user.ID, user.Name)
}
