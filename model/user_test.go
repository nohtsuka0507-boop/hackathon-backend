// db/model/user_test.go
package model

import "testing"

// TestUser_Validate は Validate() メソッドの単体テスト
func TestUser_Validate(t *testing.T) {
	// テストケースを「テーブル」として定義
	testCases := []struct {
		name    string // テストの名前
		user    User   // テスト対象の入力データ
		wantErr bool   // エラーを期待する(true)か、しない(false)か
	}{
		{
			name:    "正常なケース",
			user:    User{Name: "Taro", Age: 30},
			wantErr: false, // エラーは期待しない
		},
		{
			name:    "エラー: 名前が空",
			user:    User{Name: "", Age: 30},
			wantErr: true, // エラーを期待する
		},
		{
			name:    "エラー: 名前が長すぎる (51文字)",
			user:    User{Name: "123456789012345678901234567890123456789012345678901", Age: 30},
			wantErr: true, // エラーを期待する
		},
		{
			name:    "エラー: 年齢が若すぎる (19歳)",
			user:    User{Name: "Jiro", Age: 19},
			wantErr: true, // エラーを期待する
		},
		{
			name:    "エラー: 年齢が高すぎる (81歳)",
			user:    User{Name: "Saburo", Age: 81},
			wantErr: true, // エラーを期待する
		},
		{
			name:    "正常: 境界値 (20歳)",
			user:    User{Name: "Shiro", Age: 20},
			wantErr: false, // エラーは期待しない
		},
		{
			name:    "正常: 境界値 (80歳)",
			user:    User{Name: "Goro", Age: 80},
			wantErr: false, // エラーは期待しない
		},
	}

	// 各テストケースをループで実行
	for _, tc := range testCases {
		// t.Run で、各テストをサブテストとして実行
		t.Run(tc.name, func(t *testing.T) {
			// テスト対象のメソッド（Validate）を実行
			err := tc.user.Validate()

			// 結果の検証
			// (err != nil) は「エラーが実際に発生した」
			// tc.wantErr は「エラーが発生してほしかった」
			// この2つが一致しない（XOR）なら、テスト失敗
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}