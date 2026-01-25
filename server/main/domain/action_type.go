package domain

import "time"

// 行動種別エンティティ
type ActionType struct {
	ID          int        `json:"id"`
	ActionName  string     `json:"action_name"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	CreateUser  string     `json:"create_user"`
	UpdatedAt   time.Time  `json:"updated_at"`
	UpdateUser  string     `json:"update_user"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}
