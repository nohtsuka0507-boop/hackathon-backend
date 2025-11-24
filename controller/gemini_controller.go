package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiController struct{}

func NewGeminiController() *GeminiController {
	return &GeminiController{}
}

// フロントエンドから受け取るデータ
type GenerateRequest struct {
	ProductName string `json:"product_name"`
}

// フロントエンドに返すデータ
type GenerateResponse struct {
	Description string `json:"description"`
}

func (c *GeminiController) HandleGenerate(w http.ResponseWriter, r *http.Request) {
	// 1. リクエスト(商品名)を受け取る
	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("fail: json decode, %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 2. Gemini APIクライアントを作成
	ctx := context.Background()
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Println("fail: GEMINI_API_KEY is not set")
		http.Error(w, "API Key not set", http.StatusInternalServerError)
		return
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("fail: create gemini client, %v", err)
		http.Error(w, "Failed to create Gemini client", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// 3. Geminiに生成を依頼 (モデルは gemini-1.5-flash を使用)
	model := client.GenerativeModel("gemini-2.5-flash")
	prompt := fmt.Sprintf("フリマアプリに出品する「%s」という商品の、魅力的で購買意欲をそそる短い商品説明文を100文字以内で考えてください。", req.ProductName)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Printf("fail: generate content, %v", err)
		http.Error(w, "Failed to generate content", http.StatusInternalServerError)
		return
	}

	// 4. 生成されたテキストを取り出して返す
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		// テキスト部分を取得
		if text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(GenerateResponse{Description: string(text)})
			return
		}
	}

	http.Error(w, "No content generated", http.StatusInternalServerError)
}
