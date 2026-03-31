package domain

import "time"

type Prototype struct {
	ID                int        `json:"id"`
	UserID            int        `json:"user_id"`
	Name              string     `json:"name"`
	Description       string     `json:"description"`
	ParentPrototypeID *int       `json:"parent_prototype_id,omitempty"`
	TagIDs            []int      `json:"tag_ids"`
	CreatedAt         time.Time  `json:"created_at"`
	CreateUser        string     `json:"create_user"`
	UpdatedAt         time.Time  `json:"updated_at"`
	UpdateUser        string     `json:"update_user"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty"`
}
