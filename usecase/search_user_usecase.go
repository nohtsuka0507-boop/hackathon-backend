package usecase

import (
	"db/dao"
	"db/model"
)

// SearchUserUsecase は「ユーザーを検索する」という仕事を担当
// この仕事も、"倉庫番"（UserDAO）が必要
type SearchUserUsecase struct {
	UserDAO *dao.UserDAO
}

// NewSearchUserUsecase は、新しいUsecase（シェフ）を雇う
func NewSearchUserUsecase(userDAO *dao.UserDAO) *SearchUserUsecase {
	return &SearchUserUsecase{UserDAO: userDAO}
}

// Execute は「ユーザー検索」の仕事を実行する
//
// 戻り値は、見つかった []model.User（Userのスライス）と、エラー
func (uc *SearchUserUsecase) Execute(name string) ([]model.User, error) {
	// このUsecaseはシンプルです
	// "倉庫番"に、DBからの検索（SELECT）を指示するだけ
	users, err := uc.UserDAO.FindUsersByName(name)
	if err != nil {
		return nil, err
	}

	return users, nil
}