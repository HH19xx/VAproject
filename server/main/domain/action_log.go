package domain

import "time"

// 行動記録エンティティ
type ActionLog struct {
	ID         int        `json:"id"`
	TargetID   *int       `json:"target_id,omitempty"`
	ActionType int        `json:"action_type"`
	Timestamp  time.Time  `json:"timestamp"`
	Notes      string     `json:"notes"`
	CreatedAt  time.Time  `json:"created_at"`
	CreateUser string     `json:"create_user"`
	UpdatedAt  time.Time  `json:"updated_at"`
	UpdateUser string     `json:"update_user"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}
