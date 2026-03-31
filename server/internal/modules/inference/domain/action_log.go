package domain

import "time"

type ActionLogAttribute struct {
	Key         string   `json:"key"`
	ValueNumber *float64 `json:"value_number,omitempty"`
}

type ActionLog struct {
	ID          int                  `json:"id"`
	UserID      int                  `json:"user_id"`
	Title       string               `json:"title"`
	PrototypeID *int                 `json:"prototype_id,omitempty"`
	OccurredAt  time.Time            `json:"occurred_at"`
	Notes       string               `json:"notes"`
	TagIDs      []int                `json:"tag_ids"`
	Attributes  []ActionLogAttribute `json:"attributes,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
	CreateUser  string               `json:"create_user"`
	UpdatedAt   time.Time            `json:"updated_at"`
	UpdateUser  string               `json:"update_user"`
	DeletedAt   *time.Time           `json:"deleted_at,omitempty"`
}
