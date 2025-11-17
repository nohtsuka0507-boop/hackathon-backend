package usecase

import (
	// ↓ さっき作った2つのパッケージをインポート
	"db/dao"
	"db/model"
)

// RegisterUserUsecase は「ユーザーを登録する」という仕事を担当
// 仕事道具として、"倉庫番"（UserDAO）が必要
type RegisterUserUsecase struct {
	UserDAO *dao.UserDAO
}

// NewRegisterUserUsecase は、新しいUsecase（シェフ）を雇う
// main.goが、仕事道具（userDAO）を渡す
func NewRegisterUserUsecase(userDAO *dao.UserDAO) *RegisterUserUsecase {
	return &RegisterUserUsecase{UserDAO: userDAO}
}

// Execute は「ユーザー登録」の仕事を実行する
//
// 戻り値は、新しく作成されたUserのID（string）と、エラー
func (uc *RegisterUserUsecase) Execute(name string, age int) (string, error) {
	// 1. modelのロジックを呼び出す
	// model.NewUserが、IDの生成（ulid）とバリデーション（Validate）を両方行う
	user, err := model.NewUser(name, age)
	if err != nil {
		// (例: バリデーションエラー)
		return "", err
	}

	// 2. daoのロジックを呼び出す
	// "倉庫番"に、DBへの保存（トランザクションやINSERT）を指示する
	if err := uc.UserDAO.CreateUser(user); err != nil {
		// (例: DBへの書き込みエラー)
		return "", err
	}

	// 3. 成功したら、新しいIDを返す
	// このIDは、controllerがJSONレスポンスとして使う
	return user.Id, nil
}
