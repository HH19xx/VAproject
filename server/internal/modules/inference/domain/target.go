package domain

import "time"

type Target struct {
	ID            int        `json:"id"`
	UserID        int        `json:"user_id"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	MatchMode     string     `json:"match_mode"`
	QueryText     string     `json:"query_text"`
	TagIDs        []int      `json:"tag_ids"`
	AnyTagIDs     []int      `json:"any_tag_ids"`
	AnyTagGroups  [][]int    `json:"any_tag_groups"`
	ExcludeTagIDs []int      `json:"exclude_tag_ids"`
	CreatedAt     time.Time  `json:"created_at"`
	CreateUser    string     `json:"create_user"`
	UpdatedAt     time.Time  `json:"updated_at"`
	UpdateUser    string     `json:"update_user"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
}
