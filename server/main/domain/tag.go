package domain

import "time"

// Tag は行動記録や観察対象に付与するタグを表す
type Tag struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id"`
	Name        string     `json:"name"`
	GroupName   string     `json:"group"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	CreateUser  string     `json:"create_user"`
	UpdatedAt   time.Time  `json:"updated_at"`
	UpdateUser  string     `json:"update_user"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}
