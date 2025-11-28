package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func main() {
	godotenv.Load()
	ctx := context.Background()
	apiKey := os.Getenv("GEMINI_API_KEY")

	if apiKey == "" {
		log.Fatal("APIキーが .env から読み込めませんでした")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	fmt.Println("--- あなたが使えるAIモデル一覧 ---")
	iter := client.ListModels(ctx)
	for {
		m, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		// 画像が使えるモデルだけ表示
		if m.Name == "models/gemini-1.5-flash" || m.Name == "models/gemini-pro-vision" || m.Name == "models/gemini-1.5-flash-001" {
			fmt.Printf("★ 推奨: %s\n", m.Name)
		} else {
			fmt.Println(m.Name)
		}
	}
	fmt.Println("------------------------------")
}
