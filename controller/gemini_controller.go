package controller

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hackathon-backend/dao"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type GeminiController struct {
	ItemDAO *dao.ItemDAO
}

func NewGeminiController(itemDAO *dao.ItemDAO) *GeminiController {
	return &GeminiController{ItemDAO: itemDAO}
}

// リクエスト構造体
type GeminiRequest struct {
	Contents         []Content        `json:"contents"`
	GenerationConfig GenerationConfig `json:"generationConfig"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text       string      `json:"text,omitempty"`
	InlineData *InlineData `json:"inlineData,omitempty"`
}

type InlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type GenerationConfig struct {
	ResponseMimeType string `json:"responseMimeType"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// 共通: Gemini API呼び出し
func (c *GeminiController) callGeminiAPI(promptText string, imageData []byte, mimeType string) (string, error) {
	apiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	if apiKey == "" {
		log.Println("【致命的エラー】GEMINI_API_KEY が環境変数に設定されていません。")
		return "", fmt.Errorf("API Key is missing")
	}

	modelName := "gemini-2.5-flash"
	url := "https://generativelanguage.googleapis.com/v1beta/models/" + modelName + ":generateContent?key=" + apiKey

	maskedKey := apiKey
	if len(apiKey) > 8 {
		maskedKey = apiKey[:4] + "...." + apiKey[len(apiKey)-4:]
	}
	log.Printf("Gemini Request (%s) Start. Key: %s", modelName, maskedKey)

	parts := []Part{{Text: promptText}}
	if len(imageData) > 0 {
		base64Data := base64.StdEncoding.EncodeToString(imageData)
		parts = append(parts, Part{
			InlineData: &InlineData{
				MimeType: mimeType,
				Data:     base64Data,
			},
		})
	}

	reqBody := GeminiRequest{
		Contents: []Content{{Parts: parts}},
		GenerationConfig: GenerationConfig{
			ResponseMimeType: "application/json",
		},
	}

	jsonData, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	logBody := strings.ReplaceAll(string(bodyBytes), "\n", " ")

	if resp.StatusCode != 200 {
		log.Printf("Gemini API Error (%d): %s", resp.StatusCode, logBody)
		return "", fmt.Errorf("API Error (%d): %s", resp.StatusCode, logBody)
	}

	var geminiResp GeminiResponse
	if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&geminiResp); err != nil {
		return "", err
	}

	if geminiResp.Error.Code != 0 {
		return "", fmt.Errorf("API Error: %s", geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		return geminiResp.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("no response from AI")
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
	prompt := fmt.Sprintf("商品名「%s」の魅力的で簡潔な商品説明文を、日本語で200文字以内で書いてください。Markdownは使わず、JSON形式 {\"description\": \"...\"} で返してください。", req.ProductName)
	result, err := c.callGeminiAPI(prompt, nil, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(result))
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

	mimeType := "image/jpeg"
	filename := strings.ToLower(header.Filename)
	if strings.HasSuffix(filename, ".png") {
		mimeType = "image/png"
	} else if strings.HasSuffix(filename, ".webp") {
		mimeType = "image/webp"
	} else if strings.HasSuffix(filename, ".heic") {
		mimeType = "image/heic"
	}

	var promptText string
	if mode == "repair" {
		// ★ここを強化！AIに「利益」や「難易度」まで計算させるプロンプトに変更
		promptText = `あなたはプロのリペア職人兼、フリマアプリの相場師です。
アップロードされた画像の商品の状態を分析し、以下のJSON形式でのみ回答してください。Markdownは不要です。

{
  "item_name": "商品の推測名（例: 本革のビジネスバッグ）",
  "damage_check": "破損箇所の具体的な指摘（例: ハンドルの付け根が千切れている、角擦れがある）",
  "repair_plan": "具体的な修理手順（例: 1.革用接着剤で仮止め 2.麻糸で補強縫い 3.コバコートを塗る）",
  "difficulty": "修理難易度（5段階評価の数字のみ。例: 3）",
  "required_tools": ["必要な道具1", "必要な道具2", "必要な道具3"],
  "current_value": 現在の状態でのメルカリ想定販売価格（数値のみ。例: 1500）,
  "future_value": 修理して綺麗にした場合のメルカリ想定販売価格（数値のみ。例: 6000）,
  "estimated_profit": future_value - current_value の計算結果（数値のみ。例: 4500）,
  "advice": "高く売るためのワンポイントアドバイス（例: 写真を撮るときは自然光で、傷が目立たない角度を探しましょう）"
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

	result, err := c.callGeminiAPI(promptText, imageData, mimeType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cleanTxt := strings.ReplaceAll(result, "```json", "")
	cleanTxt = strings.ReplaceAll(cleanTxt, "```", "")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(cleanTxt))
}

// チャットチェック
func (c *GeminiController) HandleCheckContent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	prompt := fmt.Sprintf(`以下のメッセージが「攻撃的」「暴力的」か判定してください。問題あれば "UNSAFE"、なければ "SAFE" とJSONの {"result": "SAFE"} 形式で答えてください。メッセージ: "%s"`, req.Content)
	result, err := c.callGeminiAPI(prompt, nil, "")
	isSafe := true
	if err == nil {
		if strings.Contains(result, "UNSAFE") {
			isSafe = false
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"is_safe": isSafe})
}
