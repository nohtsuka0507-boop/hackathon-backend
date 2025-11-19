package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mysqlUser := os.Getenv("MYSQL_USER")
		mysqlPwd := os.Getenv("MYSQL_PWD")
		mysqlHost := os.Getenv("MYSQL_HOST")
		mysqlDatabase := os.Getenv("MYSQL_DATABASE")

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "--- Cloud Run 環境変数 診断テスト ---\n\n")
		fmt.Fprintf(w, "1. MYSQL_USER: [%s]\n", mysqlUser)
		fmt.Fprintf(w, "2. MYSQL_PWD: [%s]\n", maskPassword(mysqlPwd))
		fmt.Fprintf(w, "3. MYSQL_HOST: [%s]\n", mysqlHost)
		fmt.Fprintf(w, "4. MYSQL_DATABASE: [%s]\n", mysqlDatabase)
	})

	log.Println("Listening on :8000 ...")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}

func maskPassword(pwd string) string {
	if len(pwd) > 0 { return "******" }
	return "（空っぽ）"
}
