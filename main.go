package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"

    // --- ↓ 既存のパッケージをインポート (あなたのプロジェクトに合わせて調整) ↓ ---
    "db/controller"
    "db/dao"
    "db/usecase"
    // --- ↑ 既存のパッケージをインポート ↑ ---

    _ "github.com/go-sql-driver/mysql"
)

// main関数はアプリケーションのエントリーポイントです
func main() {
    // --- 1. DB接続 (Cloud Runの環境変数を読み込む) ---
    mysqlUser := os.Getenv("MYSQL_USER")
    mysqlPwd := os.Getenv("MYSQL_PWD")
    mysqlHost := os.Getenv("MYSQL_HOST")
    mysqlDatabase := os.Getenv("MYSQL_DATABASE")

    // Cloud Runの環境変数を使用して接続文字列を作成
    // MYSQL_HOSTに "unix(/cloudsql/...)" が入ることで、Cloud SQLに接続する
    connStr := fmt.Sprintf("%s:%s@%s/%s", mysqlUser, mysqlPwd, mysqlHost, mysqlDatabase)
    
    // DB接続を開く
    db, err := sql.Open("mysql", connStr)
    if err != nil {
        // 接続失敗は致命的なのでログに出力して終了
        log.Fatalf("fail: sql.Open, %v\n", err)
    }
    
    // 接続の確認（ここでクラッシュする可能性が高い）
    if err := db.Ping(); err != nil {
        // Ping失敗は致命的なのでログに出力して終了
        log.Fatalf("fail: db.Ping, %v\n", err)
    }
    log.Println("success: Database connection established.")

    // DB接続を安全に閉じるための defer
    defer func() {
        if err := db.Close(); err != nil {
            log.Printf("fail: db.Close(), %v\n", err)
        }
        log.Println("success: db.Close()")
    }()
    
    // シグナルハンドリング
    handleSysCall(db)

    // --- 2. 部品の組み立て（依存性の注入）---
    // DAO -> Usecase -> Controller の順に依存性を注入
    userDAO := dao.NewUserDAO(db)
    searchUserUsecase := usecase.NewSearchUserUsecase(userDAO)
    registerUserUsecase := usecase.NewRegisterUserUsecase(userDAO)
    searchUserController := controller.NewSearchUserController(searchUserUsecase)
    registerUserController := controller.NewRegisterUserController(registerUserUsecase)

    // --- 3. HTTPルーティング（リクエストの振り分け）---
    http.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            searchUserController.Handle(w, r)
        case http.MethodPost:
            registerUserController.Handle(w, r)
        default:
            log.Printf("fail: HTTP Method is %s", r.Method)
            w.WriteHeader(http.StatusMethodNotAllowed)
        }
    })
    
    // ルートパスにアクセスされた場合の処理（Not Foundを返す）
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusNotFound)
        fmt.Fprintf(w, "404 Not Found: Access /user endpoint.")
    })


    // --- 4. サーバー起動 ---
    // Cloud Runは環境変数 PORT で待ち受けるポートを指定します
    port := os.Getenv("PORT")
    if port == "" {
        port = "8000" // PORTが設定されていない場合はデフォルトの8000番を使う
    }
    
    log.Printf("Listening on :%s", port)
    if err := http.ListenAndServe(":"+port, nil); err != nil {
        log.Fatal(err)
    }
}

// シグナルを受けてDB接続を閉じる処理
func handleSysCall(db *sql.DB) {
    sig := make(chan os.Signal, 1)
    // SIGTERM (Cloud Run停止シグナル) と SIGINT を監視
    signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
    go func() {
        s := <-sig
        log.Printf("received syscall, %v", s)
        // db.Close() は defer で実行されるため、ここでは終了シグナルだけ送る
        os.Exit(0)
    }()
}
