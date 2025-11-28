package controller

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ★ メモリ上でユーザーを管理するための簡易データベース
var (
	mockUsersStore = map[string]string{} // email -> passwordHash
	mockUserInfo   = map[string]struct { // email -> UserInfo
		ID   int64
		Name string
	}{}
	mockUserIDCounter int64      = 1
	storeMutex        sync.Mutex // 同時アクセス対策
)

type AuthController struct {
	db *sql.DB
}

func NewAuthController(db *sql.DB) *AuthController {
	return &AuthController{db: db}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// ★ サインアップ（メモリに保存）
func (c *AuthController) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	storeMutex.Lock()
	defer storeMutex.Unlock()

	// すでに登録済みかチェック
	if _, exists := mockUsersStore[req.Email]; exists {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	// パスワードハッシュ化
	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	// メモリに保存
	mockUsersStore[req.Email] = string(hashedPwd)
	mockUserInfo[req.Email] = struct {
		ID   int64
		Name string
	}{ID: mockUserIDCounter, Name: req.Name}

	newID := mockUserIDCounter
	mockUserIDCounter++

	fmt.Printf("✅ 新規ユーザー登録: %s (ID: %d)\n", req.Name, newID)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": newID, "message": "User registered (Mock)"})
}

// ★ ログイン（メモリから照合）
func (c *AuthController) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	storeMutex.Lock()
	defer storeMutex.Unlock()

	hashedPassword, exists := mockUsersStore[req.Email]
	if !exists {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// パスワード照合
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	// ユーザー情報取得
	userInfo := mockUserInfo[req.Email]

	// トークン発行
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userInfo.ID,
		Email:  req.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtKey := []byte(os.Getenv("JWT_SECRET_KEY"))
	if len(jwtKey) == 0 {
		jwtKey = []byte("secret_key_for_hackathon")
	}

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Token generation failed", http.StatusInternalServerError)
		return
	}

	fmt.Printf("🔓 ログイン成功: %s\n", userInfo.Name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": tokenString,
		"user": map[string]interface{}{
			"id":    userInfo.ID,
			"name":  userInfo.Name,
			"email": req.Email,
		},
	})
}
