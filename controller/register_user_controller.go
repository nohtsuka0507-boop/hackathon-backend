package controller

import (
	"encoding/json"
	"hackathon-backend/usecase"
	"io"
	"log"
	"net/http"
)

// HTTPリクエストの期待する形を修正（Ageを削除、Email/Passwordを追加）
type UserReqForHTTPPost struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResForHTTPPost struct {
	Id string `json:"id"`
}

type RegisterUserController struct {
	Usecase *usecase.RegisterUserUsecase
}

func NewRegisterUserController(usecase *usecase.RegisterUserUsecase) *RegisterUserController {
	return &RegisterUserController{Usecase: usecase}
}

func (c *RegisterUserController) Handle(w http.ResponseWriter, r *http.Request) {
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

	// 修正: 引数を Name, Email, Password の3つに変更
	newID, err := c.Usecase.Execute(req.Name, req.Email, req.Password)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res := UserResForHTTPPost{Id: newID}
	bytes, err := json.Marshal(res)
	if err != nil {
		log.Printf("fail: json.Marshal, %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(bytes)
}
