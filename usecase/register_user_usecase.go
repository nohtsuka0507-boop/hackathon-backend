package usecase

import (
	"hackathon-backend/dao"
	"hackathon-backend/model"
)

type RegisterUserUsecase struct {
	dao *dao.UserDAO
}

func NewRegisterUserUsecase(dao *dao.UserDAO) *RegisterUserUsecase {
	return &RegisterUserUsecase{dao: dao}
}

// Execute はユーザー登録のビジネスロジックを行います
func (u *RegisterUserUsecase) Execute(name, email, password string) (string, error) {
	// 1. モデルの生成
	newUser, err := model.NewUser(name, email, password)
	if err != nil {
		return "", err
	}

	// 2. DBへの保存
	// ★修正: ここを Insert から CreateUser に変更しました
	if err := u.dao.CreateUser(newUser); err != nil {
		return "", err
	}

	// 3. IDを返す
	return newUser.ID, nil
}
