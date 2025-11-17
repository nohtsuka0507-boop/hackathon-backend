package model

import (
	"errors"
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

// User はアプリケーションの核となるモデル
// あなたのコードに合わせて、IDはstring型にします
type User struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// NewUser は新しいUserモデルを作成し、IDを生成します
// createUserHandlerからID生成ロジックを移植
func NewUser(name string, age int) (*User, error) {
	entropy := ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
	newID := ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()

	user := &User{
		Id:   newID,
		Name: name,
		Age:  age,
	}

	// すぐにバリデーションを実行
	if err := user.Validate(); err != nil {
		return nil, err
	}
	return user, nil
}

// Validate はユーザーの入力値をチェックします
// createUserHandlerからバリデーションロジックを移植
func (u *User) Validate() error {
	if u.Name == "" || len(u.Name) > 50 {
		return errors.New("fail: name is invalid")
	}
	if u.Age < 20 || u.Age > 80 {
		return errors.New("fail: age is invalid")
	}
	return nil
}