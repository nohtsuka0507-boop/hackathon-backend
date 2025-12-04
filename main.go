package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"

	"hackathon-backend/controller"
	"hackathon-backend/dao"
	"hackathon-backend/usecase"
)

func main() {

	// ログを目立たせて更新確認しやすくします
	log.Println("🔥🔥🔥 UPDATED VERSION: Like Feature Added 🔥🔥🔥")

	// --- 0. 環境変数の読み込み ---
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found")
	}

	// --- 1. DB接続 ---
	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlPwd := os.Getenv("MYSQL_PWD")
	mysqlHost := os.Getenv("MYSQL_HOST")
	mysqlDatabase := os.Getenv("MYSQL_DATABASE")

	connStr := fmt.Sprintf("%s:%s@%s/%s?parseTime=true", mysqlUser, mysqlPwd, mysqlHost, mysqlDatabase)
	log.Printf("Connecting to DB: %s@%s/%s", mysqlUser, mysqlHost, mysqlDatabase)

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Fatalf("Fatal: Failed to open DB connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Fatal: Failed to connect to Cloud SQL: %v. Check MYSQL_HOST env var!", err)
	}
	log.Println("Success: Connected to Cloud SQL!")

	// --- 1.5 テーブルの自動作成 ---
	if err := createTables(db); err != nil {
		log.Fatalf("Fatal: Failed to create tables: %v", err)
	}
	defer db.Close()

	// --- 2. 依存関係の注入 (DI) ---

	// DAOの初期化
	userDAO := dao.NewUserDAO(db)
	itemDAO := dao.NewItemDAO(db)
	messageDAO := dao.NewMessageDAO(db)
	likeDAO := dao.NewLikeDAO(db) // ★追加: いいね用DAO

	// コントローラー・ユースケースの初期化
	authController := controller.NewAuthController(userDAO)

	searchUserUsecase := usecase.NewSearchUserUsecase(userDAO)
	registerUserUsecase := usecase.NewRegisterUserUsecase(userDAO)
	searchUserController := controller.NewSearchUserController(searchUserUsecase)
	registerUserController := controller.NewRegisterUserController(registerUserUsecase)

	itemController := controller.NewItemController(itemDAO)
	geminiController := controller.NewGeminiController(itemDAO)
	chatController := controller.NewChatController(messageDAO)
	likeController := controller.NewLikeController(likeDAO) // ★追加: いいね用コントローラー

	// --- 3. ルーティング設定 ---
	mux := http.NewServeMux()

	// 認証
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authController.HandleRegister(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authController.HandleLogin(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// ユーザー
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			searchUserController.Handle(w, r)
		case http.MethodPost:
			registerUserController.Handle(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// 商品
	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			itemController.HandleGetItems(w, r)
		case http.MethodPost:
			itemController.HandleAddItem(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/items/purchase", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			itemController.HandlePurchase(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// AI
	mux.HandleFunc("/generate-description", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			geminiController.HandleGenerate(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/analyze-image", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			geminiController.HandleAnalyzeImage(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/analyze-listing", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			geminiController.HandleAnalyzeListing(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// チャット
	mux.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			chatController.HandleGetMessages(w, r)
		case http.MethodPost:
			chatController.HandlePostMessage(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// ★追加: いいね機能
	mux.HandleFunc("/likes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			// いいねの切り替え (ON/OFF)
			likeController.HandleToggleLike(w, r)
		case http.MethodGet:
			// 自分がいいねした商品一覧を取得
			likeController.HandleGetLikes(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// --- 4. サーバー起動 ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// CORSミドルウェアを適用
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: enableCORS(mux),
	}

	go func() {
		log.Printf("🚀 Server is running on http://localhost:%s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
}

// enableCORS: CORS設定 (変更なし)
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func createTables(db *sql.DB) error {
	// Itemテーブル
	queryItem := `
    CREATE TABLE IF NOT EXISTS items (
        id VARCHAR(255) PRIMARY KEY,
        name VARCHAR(255),
        price INT,
        description TEXT,
        sold_out BOOLEAN DEFAULT FALSE,
        image_url LONGTEXT
    );`
	if _, err := db.Exec(queryItem); err != nil {
		return fmt.Errorf("create items table error: %w", err)
	}

	// Userテーブル
	queryUser := `
    CREATE TABLE IF NOT EXISTS users (
        id VARCHAR(255) PRIMARY KEY,
        email VARCHAR(255) UNIQUE,
        password VARCHAR(255),
        name VARCHAR(255)
    );`
	if _, err := db.Exec(queryUser); err != nil {
		return fmt.Errorf("create users table error: %w", err)
	}

	// チャットメッセージテーブル
	queryMsg := `
    CREATE TABLE IF NOT EXISTS messages (
        id VARCHAR(255) PRIMARY KEY,
        item_id VARCHAR(255),
        sender_id VARCHAR(255),
        content TEXT,
        created_at VARCHAR(255)
    );`
	if _, err := db.Exec(queryMsg); err != nil {
		return fmt.Errorf("create messages table error: %w", err)
	}

	// ★追加: いいねテーブル
	// ユーザーIDと商品IDのペアで保存
	queryLikes := `
    CREATE TABLE IF NOT EXISTS likes (
        user_id VARCHAR(255),
        item_id VARCHAR(255),
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (user_id, item_id)
    );`
	if _, err := db.Exec(queryLikes); err != nil {
		return fmt.Errorf("create likes table error: %w", err)
	}

	return nil
}
