package domain

import "time"

// RefreshToken はリフレッシュトークンの永続化情報を表現します。
type RefreshToken struct {
    ID         int
    UserID     int
    TokenHash  string
    ExpiresAt  time.Time
    Revoked    bool
    CreatedAt  time.Time
    CreateUser string
    UpdatedAt  time.Time
    UpdateUser string
}
