package domain

type UserAuthProvider struct {
	ID             int
	UserID         int
	Provider       string
	ProviderUserID string
}
