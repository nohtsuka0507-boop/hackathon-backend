package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	// "database/sql" // 診断中はDB関連をすべて無効化
	// "os/signal"
	// "syscall"
	// "db/controller"
	// "db/dao"
	// "db/usecase"
	// _ "github.com/go-sql-driver/mysql"
)

func main() {
	// --- 1. DB接続 (クラッシュするため、診断中は一時的にすべて無効化) ---
	// mysqlUser := os.Getenv("MYSQL_USER")
	// mysqlPwd := os.Getenv("MYSQL_PWD")
	// mysqlHost := os.Getenv("MYSQL_HOST")
	// mysqlDatabase := os.Getenv("MYSQL_DATABASE")
	// ... (db.Ping() など、すべて無効)

	// --- 3. HTTPルーティング（診断用に変更）---
	// "/" (ルートURL) にアクセスが来たら、環境変数を表示する
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Cloud Runから受け取った環境変数を取得
		mysqlUser := os.Getenv("MYSQL_USER")
		mysqlPwd := os.Getenv("MYSQL_PWD")
		mysqlHost := os.Getenv("MYSQL_HOST")
		mysqlDatabase := os.Getenv("MYSQL_DATABASE")

		// 取得した値をそのままレスポンスとして書き出す
		w.Header().Set("Content-Type", "text/plain; charset=utf-8") // 文字化け防止
		fmt.Fprintf(w, "--- Cloud Run 環境変数（かんきょうへんすう） 診断テスト ---\n\n")
		fmt.Fprintf(w, "1. MYSQL_USER: [%s]\n", mysqlUser)
		fmt.Fprintf(w, "2. MYSQL_PWD (パスワード): [%s]\n", maskPassword(mysqlPwd)) // パスワードは隠します
		fmt.Fprintf(w, "3. MYSQL_HOST: [%s]\n", mysqlHost)
		fmt.Fprintf(w, "4. MYSQL_DATABASE: [%s]\n", mysqlDatabase)

		fmt.Fprintf(w, "\n\n--- 診断けっか ---\n")
		if mysqlUser == "hackathon_admin" && mysqlDatabase == "hackathon" && mysqlHost != "" && mysqlPwd == "MyNewPass2025!" {
			fmt.Fprintf(w, "成功: すべての変数がGoLandでの設定と一致しています。\n")
			fmt.Fprintf(w, "（この画面が出たままなら、元のmain.goのdb.Ping()が別の理由で失敗しています）\n")
		} else {
			fmt.Fprintf(w, "エラー: 変数が間違っているか、空っぽです。\n")
			fmt.Fprintf(w, "上の [ ] の中身と、GCPの「変数の設定」を見くらべてください。\n")
		}
	})

	// --- 4. サーバー起動 ---
	log.Println("Listening on :8000 ... (DIAGNOSTIC MODE)")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}

// パスワードを隠すためのヘルパー関数
func maskPassword(pwd string) string {
	if len(pwd) == 0 {
		return "（空っぽです）"
	}
	if len(pwd) > 0 {
		return "****** (設定されています)"
	}
	return ""
}

// (handleSysCallはDBを使わないので無効化)
// func handleSysCall(db *sql.DB) {
// ...
// }
