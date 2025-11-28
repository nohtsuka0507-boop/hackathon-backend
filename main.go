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
	// --- 0. 環境変数の読み込み ---
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found")
	}

	// --- 1. DB接続 (エラーでも止まらないように修正) ---
	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlPwd := os.Getenv("MYSQL_PWD")
	mysqlHost := os.Getenv("MYSQL_HOST")
	mysqlDatabase := os.Getenv("MYSQL_DATABASE")

	connStr := fmt.Sprintf("%s:%s@%s/%s", mysqlUser, mysqlPwd, mysqlHost, mysqlDatabase)
	db, err := sql.Open("mysql", connStr)

	// ★ 修正ポイント: DBエラーでも Fatal（強制終了）にしない
	if err != nil {
		log.Printf("Warning: DB init failed: %v (Running in No-DB Mode)\n", err)
	} else if err := db.Ping(); err != nil {
		log.Printf("Warning: DB connection failed: %v (Running in No-DB Mode)\n", err)
	} else {
		log.Println("Success: Connected to MySQL database!")
	}
	// defer db.Close() // DBがない場合のパニック防止のため削除

	// --- 2. 依存関係の注入 (DI) ---

	// ★ 認証機能
	authController := controller.NewAuthController(db)

	// ユーザー機能
	userDAO := dao.NewUserDAO(db)
	searchUserUsecase := usecase.NewSearchUserUsecase(userDAO)
	registerUserUsecase := usecase.NewRegisterUserUsecase(userDAO)
	searchUserController := controller.NewSearchUserController(searchUserUsecase)
	registerUserController := controller.NewRegisterUserController(registerUserUsecase)

	// 商品機能
	itemDAO := dao.NewItemDAO(db)
	itemController := controller.NewItemController(itemDAO)

	// AI機能
	geminiController := controller.NewGeminiController()

	// --- 3. ルーティング設定 ---
	mux := http.NewServeMux()

	// ★ 認証ルート
	mux.HandleFunc("POST /register", authController.HandleRegister)
	mux.HandleFunc("POST /login", authController.HandleLogin)

	// User Endpoints
	mux.HandleFunc("GET /user", searchUserController.Handle)
	mux.HandleFunc("POST /user", registerUserController.Handle)

	// Item Endpoints
	mux.HandleFunc("GET /items", itemController.HandleGetItems)
	mux.HandleFunc("POST /items", itemController.HandleAddItem)
	mux.HandleFunc("POST /items/purchase", itemController.HandlePurchase)

	// AI Endpoints
	mux.HandleFunc("POST /generate-description", geminiController.HandleGenerate)
	mux.HandleFunc("POST /analyze-image", geminiController.HandleAnalyzeImage)
	// ★ 出品用AI分析への道
	mux.HandleFunc("POST /analyze-listing", geminiController.HandleAnalyzeListing)

	// --- 4. サーバー起動 ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

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

	// 終了シグナル待機
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
}

// CORS設定
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
