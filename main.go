package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	// --- ↓ 既存のパッケージをすべてインポート ↓ ---
	"db/controller"
	"db/dao"
	"db/usecase"
	// --- ↑ 既存のパッケージをすべてインポート ↑ ---

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

	// --- 2. 部品の組み立て ---
	userDAO := dao.NewUserDAO(db)
	searchUserUsecase := usecase.NewSearchUserUsecase(userDAO)
	registerUserUsecase := usecase.NewRegisterUserUsecase(userDAO)
	searchUserController := controller.NewSearchUserController(searchUserUsecase)
	registerUserController := controller.NewRegisterUserController(registerUserUsecase)

	// --- 3. HTTPルーティング（CORS対応を追加！）---
	http.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		// ▼▼▼ ここからCORS許可設定 ▼▼▼
		// 誰からのアクセスでも許可する（*）
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// 許可するメソッド
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		// 許可するヘッダー
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// "OPTIONS"メソッド（ブラウザからの事前確認）が来たら、OKだけ返して終了
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		// ▲▲▲ ここまでCORS許可設定 ▲▲▲

		switch r.Method {
		case http.MethodGet:
			searchUserController.Handle(w, r)
		case http.MethodPost:
			registerUserController.Handle(w, r)
		default:
			log.Printf("fail: HTTP Method is %s\n", r.Method)
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

// システムコールハンドラ
func handleSysCall(db *sql.DB) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sig
		log.Printf("received syscall, %v", s)
		os.Exit(0)
	}()
}
