package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	// --- ↓ 既存のパッケージをインポート ↓ ---
	"db/controller"
	"db/dao"
	"db/usecase"
	// --- ↑ 既存のパッケージをインポート ↑ ---

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// --- 1. DB接続 (Cloud Runの環境変数を読み込むように修正) ---
	// 資料(14.39.47.jpg)のコードブロックを適用
	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlPwd := os.Getenv("MYSQL_PWD")
	mysqlHost := os.Getenv("MYSQL_HOST")
	mysqlDatabase := os.Getenv("MYSQL_DATABASE")

	// 資料(14.39.47.jpg)の接続文字列(connStr)の形式に変更
	// MYSQL_HOSTに "unix(/cloudsql/...)" が入ることで、Cloud SQLに接続する
	connStr := fmt.Sprintf("%s:%s@%s/%s", mysqlUser, mysqlPwd, mysqlHost, mysqlDatabase)
	db, err := sql.Open("mysql", connStr) // dsn -> connStr に変更
	if err != nil {
		log.Fatalf("fail: sql.Open, %v\n", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("fail: _db.Ping, %v\n", err)
	}

	// アプリケーション終了時にDB接続を閉じる
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("fail: db.Close(), %v\n", err)
		}
		log.Println("success: db.Close()")
	}()
	handleSysCall(db) // dbを渡す

	// --- 2. 部品の組み立て（依存性の注入）---
	// (この部分は元のコードのまま)
	userDAO := dao.NewUserDAO(db)
	searchUserUsecase := usecase.NewSearchUserUsecase(userDAO)
	registerUserUsecase := usecase.NewRegisterUserUsecase(userDAO)
	searchUserController := controller.NewSearchUserController(searchUserUsecase)
	registerUserController := controller.NewRegisterUserController(registerUserUsecase)

	// --- 3. HTTPルーティング（リクエストの振り分け）---
	// (この部分は元のコードのまま)
	http.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
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
	// Cloud Runは自動でPORT環境変数を設定するが、
	// ここでは資料の指示にないため、元のコードの "8000" のままにします。
	// (Cloud Runは8000ポートにリクエストを送ってくれるので問題ありません)
	log.Println("Listening on :8000 ...")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}

// (この関数は元のコードのまま)
func handleSysCall(db *sql.DB) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sig
		log.Printf("received syscall, %v", s)
		os.Exit(0)
	}()
}
