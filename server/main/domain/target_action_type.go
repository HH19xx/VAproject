package domain

import "time"

// TargetActionType は観察対象と行動種別の紐づけを表す
type TargetActionType struct {
	ID           int       `json:"id"`
	TargetID     int       `json:"target_id"`
	ActionTypeID int       `json:"action_type_id"`
	CreatedAt    time.Time `json:"created_at"`
	CreateUser   string    `json:"create_user"`
}
