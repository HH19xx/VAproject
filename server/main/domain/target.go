package domain

import "time"

// 観察対象のエンティティ。
// ユーザーが観察・記録したい対象（人物、動物、物など）を表す。
type Target struct {
	ID          int        // 観察対象の一意な識別子
	Name        string     // 観察対象の名前（最大64文字、UNIQUE制約）
	Description string     // 観察対象の説明（任意、TEXT型）
	CreatedAt   time.Time  // レコード作成日時
	CreateUser  string     // 作成者ユーザー名（最大32文字）
	UpdatedAt   time.Time  // レコード最終更新日時
	UpdateUser  string     // 更新者ユーザー名（最大32文字）
	DeletedAt   *time.Time // 論理削除日時（NULLなら削除されていない）
}
