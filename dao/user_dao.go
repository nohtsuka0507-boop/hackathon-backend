package dao

import (
	"database/sql"
	"hackathon-backend/model"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

type UserDAO struct {
	DB *sql.DB
}

func NewUserDAO(db *sql.DB) *UserDAO {
	return &UserDAO{DB: db}
}

// CreateUser は、ユーザーをDB(usersテーブル)に登録します
func (d *UserDAO) CreateUser(user *model.User) error {
	// main.go で定義したテーブル名 "users" に合わせます
	// カラム: id, name, email, password
	query := "INSERT INTO users (id, name, email, password) VALUES (?, ?, ?, ?)"

	// シンプルなExecで十分ですが、元のコードを尊重してトランザクションを使う形でもOKです。
	// ここでは確実に動くシンプルな形にします。
	_, err := d.DB.Exec(query, user.ID, user.Name, user.Email, user.Password)
	if err != nil {
		log.Printf("fail: insert user, %v\n", err)
		return err
	}
	return nil
}

// GetUserByEmail は、メールアドレスからユーザーを取得します（ログイン用）
func (d *UserDAO) GetUserByEmail(email string) (*model.User, error) {
	user := &model.User{}
	query := "SELECT id, name, email, password FROM users WHERE email = ?"

	row := d.DB.QueryRow(query, email)
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // ユーザーが見つからない場合
		}
		log.Printf("fail: get user by email, %v\n", err)
		return nil, err
	}

	return user, nil
}

// FindUsersByName は、Nameでユーザーを検索します（既存機能の維持）
func (d *UserDAO) FindUsersByName(name string) ([]model.User, error) {
	// テーブル名を "users" に修正
	// Age はなくなったので取得しません
	query := "SELECT id, name, email, password FROM users WHERE name = ?"

	rows, err := d.DB.Query(query, name)
	if err != nil {
		log.Printf("fail: db.Query, %v\n", err)
		return nil, err
	}
	defer rows.Close()

	users := make([]model.User, 0)
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Password); err != nil {
			log.Printf("fail: rows.Scan, %v\n", err)
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}
