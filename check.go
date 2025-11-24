package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()

	// ★ここに確実に正しいキーを入れてください
	apiKey := "AIzaSyDonVixyHLzxz8ZeGDsP_rgmjYm2OSRhmY"

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// 使えるモデルを全部リストアップして表示する
	iter := client.ListModels(ctx)
	fmt.Println("--- あなたのAPIキーで使えるモデル一覧 ---")
	for {
		m, err := iter.Next()
		if err != nil {
			break
		}
		fmt.Println(m.Name)
	}
	fmt.Println("---------------------------------------")
}
