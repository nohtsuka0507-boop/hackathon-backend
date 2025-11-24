package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"hackathon-backend/controller"
	"hackathon-backend/dao"
	"hackathon-backend/usecase"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// --- 1. DB接続 ---
	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlPwd := os.Getenv("MYSQL_PWD")
	mysqlHost := os.Getenv("MYSQL_HOST")
	mysqlDatabase := os.Getenv("MYSQL_DATABASE")

	connStr := fmt.Sprintf("%s:%s@%s/%s", mysqlUser, mysqlPwd, mysqlHost, mysqlDatabase)
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Fatalf("fail: sql.Open, %v\n", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("fail: _db.Ping, %v\n", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("fail: db.Close(), %v\n", err)
		}
	}()
	handleSysCall(db)

	// --- 2. コントローラーの初期化 ---

	// User関連
	userDAO := dao.NewUserDAO(db)
	searchUserUsecase := usecase.NewSearchUserUsecase(userDAO)
	registerUserUsecase := usecase.NewRegisterUserUsecase(userDAO)
	searchUserController := controller.NewSearchUserController(searchUserUsecase)
	registerUserController := controller.NewRegisterUserController(registerUserUsecase)

	// Gemini関連
	geminiController := controller.NewGeminiController()

	// ★ Item関連 (ここを追加)
	itemDAO := dao.NewItemDAO(db)
	itemController := controller.NewItemController(itemDAO)

	// --- 3. ルーティング設定 ---

	// Userルーティング
	http.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		switch r.Method {
		case http.MethodGet:
			searchUserController.Handle(w, r)
		case http.MethodPost:
			registerUserController.Handle(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Geminiルーティング
	http.HandleFunc("/generate-description", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == http.MethodPost {
			geminiController.HandleGenerate(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// ★ Itemルーティング (一覧取得・出品)
	http.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		switch r.Method {
		case http.MethodGet:
			itemController.HandleGetItems(w, r)
		case http.MethodPost:
			itemController.HandleAddItem(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// ★ Itemルーティング (購入)
	http.HandleFunc("/items/purchase", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == http.MethodPost {
			itemController.HandlePurchase(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// --- 4. サーバー起動 ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on :%s ...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handleSysCall(db *sql.DB) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sig
		log.Printf("received syscall, %v", s)
		os.Exit(0)
	}()
}
