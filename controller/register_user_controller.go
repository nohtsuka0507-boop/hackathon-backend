package controller

import (
	// ↓ さっき作った「頭脳」パッケージをインポート
	"db/usecase"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// createUserHandlerから移植するHTTPリクエスト用の構造体
type UserReqForHTTPPost struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// createUserHandlerから移植するHTTPレスポンス用の構造体
type UserResForHTTPPost struct {
	Id string `json:"id"`
}

// RegisterUserController は「ユーザー登録」の受付
// 仕事道具として、"シェフ"（RegisterUserUsecase）が必要
type RegisterUserController struct {
	Usecase *usecase.RegisterUserUsecase
}

// NewRegisterUserController は、新しいController（ウェイター）を雇う
// main.goが、仕事道具（usecase）を渡す
func NewRegisterUserController(usecase *usecase.RegisterUserUsecase) *RegisterUserController {
	return &RegisterUserController{Usecase: usecase}
}

// Handle は http.HandlerFunc として振る舞う
// createUserHandlerのロjックを丸ごと移植
func (c *RegisterUserController) Handle(w http.ResponseWriter, r *http.Request) {
	// 1. HTTPリクエストのJSONをパース（main.goから移植）
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("fail: io.ReadAll, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var req UserReqForHTTPPost
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("fail: json.Unmarshal, %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 2. "シェフ"（Usecase）に仕事を依頼
	// ControllerはバリデーションやID生成を「しない」。Usecaseに任せる。
	newID, err := c.Usecase.Execute(req.Name, req.Age)
	if err != nil {
		// Usecaseからエラーが返ってきた（バリデーションエラーなど）
		log.Println(err.Error())
		w.WriteHeader(http.StatusBadRequest) // 400 Bad Request を返す
		return
	}

	// 3. 成功レスポンスをJSONで返す（main.goから移植）
	res := UserResForHTTPPost{Id: newID}
	bytes, err := json.Marshal(res)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 登録成功なので 201 Created を返す
	w.Write(bytes)
}