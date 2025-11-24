package dao

import (
	"context"
	"database/sql"
	// さっき作った "db/model" パッケージをインポート
	"hackathon-backend/model"
	"log"

	// main.goから持ってくる
	_ "github.com/go-sql-driver/mysql"
)

// UserDAO はデータベース接続(*sql.DB)を保持します。
type UserDAO struct {
	DB *sql.DB
}

// NewUserDAO は、新しいUserDAOを初期化します。
func NewUserDAO(db *sql.DB) *UserDAO {
	return &UserDAO{DB: db}
}

// FindUsersByName は、Nameでユーザーを検索するSQLロジック（getUserHandlerから移植）
//
// 戻り値は []model.User（Userのスライス）です。
func (d *UserDAO) FindUsersByName(name string) ([]model.User, error) {
	// --- ↓ 移植したロジック ↓ ---
	rows, err := d.DB.Query("SELECT id, name, age FROM user WHERE name = ?", name)
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		return nil, err // HTTP応答はせず、エラーだけを返す
	}
	defer rows.Close()

	users := make([]model.User, 0) // model.User を使う
	for rows.Next() {
		var u model.User // model.User を使う
		if err := rows.Scan(&u.Id, &u.Name, &u.Age); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)
			return nil, err // HTTP応答はせず、エラーだけを返す
		}
		users = append(users, u)
	}
	// --- ↑ 移植したロジック ↑ ---

	return users, nil // JSON応答はせず、Userのスライスを返す
}

// CreateUser は、ユーザーをDBに登録するSQLロジック（createUserHandlerから移植）
//
// 引数は *model.User です。
func (d *UserDAO) CreateUser(user *model.User) error {
	// --- ↓ 移植したロジック ↓ ---
	// トランザクションを開始
	tx, err := d.DB.Begin()
	if err != nil {
		log.Printf("fail: db.Begin, %v\n", err)
		return err // HTTP応答はせず、エラーだけを返す
	}

	// SQLを実行
	_, err = tx.ExecContext(context.Background(), "INSERT INTO user (id, name, age) VALUES (?, ?, ?)", user.Id, user.Name, user.Age)
	if err != nil {
		log.Printf("fail: tx.ExecContext, %v\n", err)
		tx.Rollback() // エラー時はロールバック
		return err    // HTTP応答はせず、エラーだけを返す
	}

	// トランザクションをコミット
	if err := tx.Commit(); err != nil {
		log.Printf("fail: tx.Commit, %v\n", err)
		return err // HTTP応答はせず、エラーだけを返す
	}
	// --- ↑ 移植したロジック ↑ ---

	return nil // 成功
}
