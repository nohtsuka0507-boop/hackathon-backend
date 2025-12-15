package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"hackathon-backend/dao"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiController struct {
	ItemDAO *dao.ItemDAO
}

func NewGeminiController(itemDAO *dao.ItemDAO) *GeminiController {
	return &GeminiController{ItemDAO: itemDAO}
}

// HandleGenerate: テキスト生成
func (c *GeminiController) HandleGenerate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProductName string `json:"productName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	apiKey := os.Getenv("GEMINI_API_KEY")
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		http.Error(w, "AI init failed", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// ★修正: 安定版の軽量モデル gemini-1.5-flash に変更
	genModel := client.GenerativeModel("gemini-1.5-flash")

	prompt := fmt.Sprintf("商品名「%s」の魅力的で簡潔な商品説明文を、日本語で200文字以内で書いてください。Markdownは使わず、テキストのみで返してください。", req.ProductName)

	resp, err := genModel.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Printf("Gemini Gen Error: %v", err)
		http.Error(w, "AI generation failed", http.StatusInternalServerError)
		return
	}
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			response := map[string]string{"description": string(txt)}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
	}
	http.Error(w, "No response from AI", http.StatusInternalServerError)
}

// HandleAnalyzeImage: リペア診断
func (c *GeminiController) HandleAnalyzeImage(w http.ResponseWriter, r *http.Request) {
	c.analyzeImageCommon(w, r, "repair")
}

// HandleAnalyzeListing: 出品用AIアシスタント
func (c *GeminiController) HandleAnalyzeListing(w http.ResponseWriter, r *http.Request) {
	c.analyzeImageCommon(w, r, "listing")
}

// 共通処理
func (c *GeminiController) analyzeImageCommon(w http.ResponseWriter, r *http.Request, mode string) {
	// ファイルサイズ制限などを設定
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
	} else if strings.HasSuffix(filename, ".webp") {
		imageFormat = "webp"
	} else if strings.HasSuffix(filename, ".heic") {
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

	// ★修正: 安定版の軽量モデル gemini-1.5-flash に変更
	genModel := client.GenerativeModel("gemini-1.5-flash")

	// ★追加: AIに「必ずJSONで返せ」と強制する設定（パースエラー防止）
	genModel.ResponseMIMEType = "application/json"

	var promptText string
	if mode == "repair" {
		promptText = `あなたはプロのリペア職人です。画像を分析し以下のJSONスキーマに従って情報を返してください。
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
		promptText = `あなたはフリマアプリの出品代行AIです。画像を分析し、売れやすい商品情報を以下のJSONスキーマに従って返してください。
{
  "title": "キャッチーな商品名（40文字以内）",
  "description": "検索にヒットしやすい魅力的な商品説明文（200文字程度）。状態や特徴を含める。",
  "category": "最適なカテゴリ名",
  "tags": ["タグ1", "タグ2", "タグ3"],
  "suggested_price": 5000
}`
	}

	prompt := []genai.Part{
		genai.ImageData(imageFormat, imageData),
		genai.Text(promptText),
	}

	resp, err := genModel.GenerateContent(ctx, prompt...)
	if err != nil {
		log.Printf("Gemini Error: %v", err)
		http.Error(w, "AI生成エラー", http.StatusInternalServerError)
		return
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			// ResponseMIMETypeを指定したので、余計なMarkdown記法は基本的になくなりますが、念のためクリーンアップ
			cleanTxt := string(txt)
			cleanTxt = strings.ReplaceAll(cleanTxt, "```json", "")
			cleanTxt = strings.ReplaceAll(cleanTxt, "```", "")

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(cleanTxt))
			return
		}
	}
	http.Error(w, "AI応答なし", http.StatusInternalServerError)
}

// チャットの不適切発言チェック
func (c *GeminiController) HandleCheckContent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	apiKey := os.Getenv("GEMINI_API_KEY")
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		http.Error(w, "AI init failed", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// ★修正: 安定版の軽量モデル gemini-1.5-flash に変更
	genModel := client.GenerativeModel("gemini-1.5-flash")

	prompt := fmt.Sprintf(`あなたはコンテンツモデレーターです。以下のメッセージが「攻撃的」「暴力的」「差別的」「性的」な内容を含むか判定してください。

メッセージ: "%s"

判定ルール:
- 問題がある場合は "UNSAFE" とだけ答えてください。
- 問題がない場合は "SAFE" とだけ答えてください。
- 余計な説明は一切不要です。`, req.Content)

	resp, err := genModel.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Printf("Gemini Check Error: %v", err)
		// エラー時は安全側に倒して通す、あるいはエラーを返す（ここは運用によるが一旦元のまま）
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"is_safe": true})
		return
	}

	isSafe := true
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			answer := strings.TrimSpace(string(txt))
			if strings.Contains(answer, "UNSAFE") {
				isSafe = false
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"is_safe": isSafe})
}
