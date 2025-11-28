package controller

import (
	"context"
	"encoding/json" // 追加
	"fmt"           // 追加
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiController struct{}

func NewGeminiController() *GeminiController {
	return &GeminiController{}
}

// テキスト生成 (中身を実装しました)
func (c *GeminiController) HandleGenerate(w http.ResponseWriter, r *http.Request) {
	// 1. リクエストボディのJSONを読み込む
	var req struct {
		ProductName string `json:"productName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 2. AIクライアントの初期化
	ctx := context.Background()
	apiKey := os.Getenv("GEMINI_API_KEY")
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		http.Error(w, "AI init failed", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// 3. モデル選択 (gemini-2.0-flash)
	model := client.GenerativeModel("gemini-2.0-flash")

	// 4. プロンプト作成
	prompt := fmt.Sprintf("商品名「%s」の魅力的で簡潔な商品説明文を、日本語で200文字以内で書いてください。Markdownは使わず、テキストのみで返してください。", req.ProductName)

	// 5. 生成実行
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Printf("Gemini Gen Error: %v", err)
		http.Error(w, "AI generation failed", http.StatusInternalServerError)
		return
	}

	// 6. 結果を返す
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			response := map[string]string{"description": string(txt)}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(response); err != nil {
				log.Printf("JSON Encode Error: %v", err)
			}
			return
		}
	}
	http.Error(w, "No response from AI", http.StatusInternalServerError)
}

// リペア診断 (既存)
func (c *GeminiController) HandleAnalyzeImage(w http.ResponseWriter, r *http.Request) {
	c.analyzeImageCommon(w, r, "repair")
}

// ★ 新機能: 出品用AIアシスタント
func (c *GeminiController) HandleAnalyzeListing(w http.ResponseWriter, r *http.Request) {
	c.analyzeImageCommon(w, r, "listing")
}

// 共通処理
func (c *GeminiController) analyzeImageCommon(w http.ResponseWriter, r *http.Request, mode string) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "画像サイズ過大", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "画像取得失敗", http.StatusBadRequest)
		return
	}
	defer file.Close()

	imageData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "読み込み失敗", http.StatusInternalServerError)
		return
	}

	imageFormat := "jpeg"
	filename := strings.ToLower(header.Filename)
	if strings.HasSuffix(filename, ".png") {
		imageFormat = "png"
	}
	if strings.HasSuffix(filename, ".webp") {
		imageFormat = "webp"
	}
	if strings.HasSuffix(filename, ".heic") {
		imageFormat = "heic"
	}

	ctx := context.Background()
	apiKey := os.Getenv("GEMINI_API_KEY")
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		http.Error(w, "AIエラー", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash")

	var promptText string
	if mode == "repair" {
		// リペア診断用のプロンプト
		promptText = `あなたはプロのリペア職人です。画像を分析しJSONのみ返してください。Markdown不要。
{
  "item_name": "商品名",
  "damage_check": "状態",
  "repair_plan": "リペア案",
  "repair_cost": 3000,
  "current_value": 1000, 
  "future_value": 5000,
  "profit_message": "利益アップ！",
  "is_worth_repairing": true,
  "is_safe": true,
  "safety_reason": "安全"
}`
	} else {
		// ★ 出品用のプロンプト
		promptText = `あなたはフリマアプリの出品代行AIです。画像を分析し、売れやすい商品情報をJSONのみで返してください。Markdown不要。
{
  "title": "キャッチーな商品名（40文字以内）",
  "description": "検索にヒットしやすい魅力的な商品説明文（200文字程度）。状態や特徴を含める。",
  "category": "最適なカテゴリ名",
  "tags": ["タグ1", "タグ2", "タグ3", "タグ4", "タグ5"],
  "suggested_price": 5000
}`
	}

	prompt := []genai.Part{
		genai.ImageData(imageFormat, imageData),
		genai.Text(promptText),
	}

	resp, err := model.GenerateContent(ctx, prompt...)
	if err != nil {
		log.Printf("Gemini Error: %v", err)
		http.Error(w, "AI生成エラー", http.StatusInternalServerError)
		return
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			cleanTxt := strings.ReplaceAll(string(txt), "```json", "")
			cleanTxt = strings.ReplaceAll(cleanTxt, "```", "")
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(cleanTxt))
			return
		}
	}
	http.Error(w, "AI応答なし", http.StatusInternalServerError)
}
