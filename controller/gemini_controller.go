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

// HandleGenerate: 商品説明文生成
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

// 共通処理: 画像分析
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
		promptText = `あなたはプロのリペア職人兼、フリマアプリの相場師です。
アップロードされた画像の商品の状態を分析し、以下のJSON形式でのみ回答してください。Markdownは不要です。

{
  "item_name": "商品の推測名",
  "damage_check": "状態",
  "repair_plan": "具体的な修理手順",
  "difficulty": "修理難易度（1-5）",
  "required_tools": ["道具1", "道具2"],
  "current_value": 現在の状態でのメルカリ想定価格（数値）,
  "future_value": 修理後のメルカリ想定価格（数値）,
  "estimated_profit": future_value - current_value の計算結果（数値）,
  "pro_service_cost": 専門業者に修理を依頼した場合の想定費用（数値。高めに設定せよ）,
  "shipping_cost": 往復の想定送料（数値。例:1500）,
  "pro_profit": future_value - pro_service_cost - shipping_cost の計算結果（数値。マイナスになっても良い）,
  "advice": "アドバイス"
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

// HandleCheckContent: 不適切コンテンツチェック
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

// ★追加: 職人チャット機能
func (c *GeminiController) HandleCraftsmanChat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 職人「山田匠」のペルソナ定義
	prompt := fmt.Sprintf(`
あなたは「Re:Value」専属のベテランリペア職人「山田匠（やまだ たくみ）」です。
以下の設定を守って回答してください。
・一人称は「私」または「職人」、口調は丁寧だが少し職人気質な「〜ですね」「〜だと思いますよ」という温かい口調。
・ユーザーはリペア初心者です。難しそうな専門用語は避けて、100均やホームセンターで買える道具を使った解決策を提案してください。
・JSON形式 {"reply": "..."} で返答してください。

ユーザーの質問: "%s"
`, req.Message)

	result, err := c.callGeminiAPI(prompt, nil, "")
	if err != nil {
		http.Error(w, "AI Error", http.StatusInternalServerError)
		return
	}

	cleanTxt := strings.ReplaceAll(result, "```json", "")
	cleanTxt = strings.ReplaceAll(cleanTxt, "```", "")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(cleanTxt))
}
