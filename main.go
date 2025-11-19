package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	// --- ↓ 新しく作ったパッケージをすべてインポート ↓ ---
	"db/controller"
	"db/dao"
	"db/usecase"
	// --- ↑ 新しく作ったパッケージをすべてインポート ↑ ---

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// --- 1. DB接続 (Cloud Runの環境変数を読み込むように修正) ---
	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlPwd := os.Getenv("MYSQL_PWD")
	mysqlHost := os.Getenv("MYSQL_HOST") // unix(/cloudsql/...)が入ることを期待
	mysqlDatabase := os.Getenv("MYSQL_DATABASE")

	// Cloud SQLインスタンスに接続
	connStr := fmt.Sprintf("%s:%s@%s/%s", mysqlUser, mysqlPwd, mysqlHost, mysqlDatabase)
	db, err := sql.Open("mysql", connStr) 
	if err != nil {
		log.Fatalf("fail: sql.Open, %v\n", err)
	}
	if err := db.Ping(); err != nil {
		// ここでクラッシュすると、Cloud Runが仮のページを返す
		log.Fatalf("fail: _db.Ping, %v\n", err)
	}

	// アプリケーション終了時にDB接続を閉じる
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("fail: db.Close(), %v\n", err)
		}
		log.Println("success: db.Close()")
	}()
	handleSysCall(db)

	// --- 2. 部品の組み立て（依存性の注入）---
	// (DB) -> DAO -> Usecase -> Controller
	userDAO := dao.NewUserDAO(db)
	searchUserUsecase := usecase.NewSearchUserUsecase(userDAO)
	registerUserUsecase := usecase.NewRegisterUserUsecase(userDAO)
	searchUserController := controller.NewSearchUserController(searchUserUsecase)
	registerUserController := controller.NewRegisterUserController(registerUserUsecase)

	// --- 3. HTTPルーティング（リクエストの振り分け）---
	http.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// /user (GET) でユーザーリストを返す
			searchUserController.Handle(w, r)
		case http.MethodPost:
			// /user (POST) でユーザー登録を行う
			registerUserController.Handle(w, r)
		default:
			log.Printf("fail: HTTP Method is %s\n", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// --- 4. サーバー起動 ---
	// Cloud Runは8080ポートに来るリクエストを環境変数PORTで指定するため、それを採用
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // デフォルトポート
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

