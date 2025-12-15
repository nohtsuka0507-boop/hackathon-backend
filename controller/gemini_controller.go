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

// 単一のモデルで試行する関数
func (c *GeminiController) tryModel(apiKey, modelName, promptText string, imageData []byte, mimeType string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", modelName, apiKey)

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
		return "", err // 内部エラー
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err // ネットワークエラー
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	// 404の場合は「このモデルは使えない」と判断してエラーを返す
	if resp.StatusCode == 404 {
		return "", fmt.Errorf("NOT_FOUND")
	}

	// 429 (Resource Exhausted) の場合は、モデル自体は存在するので、API制限エラーとして返す
	if resp.StatusCode == 429 {
		return "", fmt.Errorf("QUOTA_EXCEEDED")
	}

	if resp.StatusCode != 200 {
		logBody := strings.ReplaceAll(string(bodyBytes), "\n", " ")
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

// 共通: Gemini API呼び出し（総当たり版）
func (c *GeminiController) callGeminiAPI(promptText string, imageData []byte, mimeType string) (string, error) {
	apiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	if apiKey == "" {
		log.Println("【致命的エラー】GEMINI_API_KEY が環境変数に設定されていません。")
		return "", fmt.Errorf("API Key is missing")
	}

	// ★総当たりリスト: 使える可能性のあるモデル名を優先順に並べる
	modelCandidates := []string{
		"gemini-1.5-flash",     // 本命
		"gemini-1.5-flash-001", // バージョン指定
		"gemini-1.5-flash-002", // 新バージョン
		"gemini-1.5-pro",       // Pro版（Flashダメならこれ）
		"gemini-1.5-pro-001",
		"gemini-pro",           // 旧バージョン (1.0)
		"gemini-2.0-flash-exp", // 実験版（元々使っていたもの）
	}

	var lastError error

	// 順番に試していく
	for _, model := range modelCandidates {
		log.Printf("Trying model: %s ...", model)
		result, err := c.tryModel(apiKey, model, promptText, imageData, mimeType)

		if err == nil {
			log.Printf("SUCCESS with model: %s", model)
			return result, nil
		}

		// エラー内容を確認
		if err.Error() == "NOT_FOUND" {
			log.Printf("Model %s not found. Skipping...", model)
			lastError = err
			continue // 次のモデルへ
		}

		// 404以外のエラー（API制限など）は、モデル自体は合っているのでそこで終了するか、続けるか判断
		// ここではとりあえずログに出して次へ（別モデルなら枠が空いてる可能性もあるため）
		log.Printf("Model %s failed with error: %v. Trying next...", model, err)
		lastError = err
	}

	// 全部ダメだった場合
	// 最後に、本当に使えるモデル一覧をログに出して死ぬ（デバッグ用）
	log.Printf("【全滅】すべてのモデルで失敗しました。利用可能なモデルを確認します...")
	logAvailableModels(apiKey)

	return "", fmt.Errorf("All models failed. Last error: %v", lastError)
}

// デバッグ用: 利用可能なモデル一覧を取得
func logAvailableModels(apiKey string) {
	url := "https://generativelanguage.googleapis.com/v1beta/models?key=" + apiKey
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to list models: %v", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	log.Printf("【DEBUG】Available Models: %s", strings.ReplaceAll(string(body), "\n", " "))
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
