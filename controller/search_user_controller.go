package controller

import (
	// "db/model" // modelはJSONの「型」として使う
	"db/usecase"
	"encoding/json"
	"log"
	"net/http"
)

// SearchUserController は「ユーザー検索」の受付
// 仕事道具として、"シェフ"（SearchUserUsecase）が必要
type SearchUserController struct {
	Usecase *usecase.SearchUserUsecase
}

// NewSearchUserController は、新しいController（ウェイター）を雇う
func NewSearchUserController(usecase *usecase.SearchUserUsecase) *SearchUserController {
	return &SearchUserController{Usecase: usecase}
}

// Handle は http.HandlerFunc として振る舞う
// getUserHandlerのロジックを丸ごと移植
func (c *SearchUserController) Handle(w http.ResponseWriter, r *http.Request) {
	// 1. HTTPリクエストのクエリパラメータをパース（main.goから移植）
	name := r.URL.Query().Get("name")
	if name == "" {
		log.Println("fail: name parameter is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 2. "シェフ"（Usecase）に仕事を依頼
	users, err := c.Usecase.Execute(name)
	if err != nil {
		// Usecaseからエラーが返ってきた（DBエラーなど）
		log.Printf(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 3. 成功レスポンスをJSONで返す（main.goから移植）
	// もしユーザーが見つからなくても、空のスライス [] を返すのが一般的
	bytes, err := json.Marshal(users)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}