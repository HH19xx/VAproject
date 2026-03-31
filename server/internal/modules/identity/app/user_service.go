package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"

	identitydomain "server/internal/modules/identity/domain"
	jwtadapter "server/internal/modules/identity/adapters/jwt"

	"golang.org/x/crypto/bcrypt"
)

var ErrAuthFailed = errors.New("authentication failed")

type UserRepository interface {
	FindByName(ctx context.Context, name string) (*identitydomain.User, error)
	FindByID(ctx context.Context, userID int) (*identitydomain.User, error)
	UpdateName(ctx context.Context, userID int, newName string) error
	FindByEmail(ctx context.Context, email string) (*identitydomain.User, error)
	Create(ctx context.Context, user *identitydomain.User) (int, error)
}

type UserAuthProviderRepository interface {
	FindByUserID(ctx context.Context, userID int) ([]*identitydomain.UserAuthProvider, error)
	HasProvider(ctx context.Context, userID int, provider string) (bool, error)
	Create(ctx context.Context, authProvider *identitydomain.UserAuthProvider) error
	Delete(ctx context.Context, userID int, provider string) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *identitydomain.RefreshToken) error
	FindByHash(ctx context.Context, hash string) (*identitydomain.RefreshToken, error)
	Revoke(ctx context.Context, id int, updateUser string) error
}

type UserService struct {
	Repo             UserRepository
	AuthProviderRepo UserAuthProviderRepository
	RefreshRepo      RefreshTokenRepository
}

type IssuedTokens struct {
	AccessToken          string
	RefreshToken         string
	AccessTokenExpiresAt time.Time
}

type UserProfile struct {
	User          *identitydomain.User
	AuthProviders []string
}

func (u *UserService) Login(ctx context.Context, name, password string) (int, error) {
	user, err := u.Repo.FindByName(ctx, name)
	if err != nil {
		return 0, ErrAuthFailed
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return 0, ErrAuthFailed
	}

	return user.ID, nil
}

func (u *UserService) Register(ctx context.Context, name, password string) (int, error) {
	existingUser, err := u.Repo.FindByName(ctx, name)
	if err == nil && existingUser != nil {
		return 0, errors.New("user name is already in use")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	newUser := &identitydomain.User{
		Name:     name,
		Password: string(hashedPassword),
	}

	userID, err := u.Repo.Create(ctx, newUser)
	if err != nil {
		return 0, err
	}

	authProvider := &identitydomain.UserAuthProvider{
		UserID:   userID,
		Provider: "local",
	}
	if err := u.AuthProviderRepo.Create(ctx, authProvider); err != nil {
		return 0, err
	}

	return userID, nil
}

func (u *UserService) GetProfile(ctx context.Context, userID int) (*UserProfile, error) {
	user, err := u.Repo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	user.Password = ""

	authProviders, err := u.AuthProviderRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("failed to fetch auth providers")
	}

	providers := make([]string, len(authProviders))
	for i, ap := range authProviders {
		providers[i] = ap.Provider
	}

	return &UserProfile{
		User:          user,
		AuthProviders: providers,
	}, nil
}

func (u *UserService) UpdateProfile(ctx context.Context, userID int, newName string) error {
	if newName == "" {
		return errors.New("user name is required")
	}

	existingUser, err := u.Repo.FindByName(ctx, newName)
	if err == nil && existingUser != nil && existingUser.ID != userID {
		return errors.New("user name is already in use")
	}

	if err := u.Repo.UpdateName(ctx, userID, newName); err != nil {
		return errors.New("failed to update user name")
	}

	return nil
}

func (u *UserService) FindOrCreateUserByEmail(ctx context.Context, email string) (int, error) {
	user, err := u.Repo.FindByEmail(ctx, email)
	if err == nil {
		hasGoogle, _ := u.AuthProviderRepo.HasProvider(ctx, user.ID, "google")
		if !hasGoogle {
			authProvider := &identitydomain.UserAuthProvider{
				UserID:   user.ID,
				Provider: "google",
			}
			u.AuthProviderRepo.Create(ctx, authProvider)
		}
		return user.ID, nil
	}

	newUser := &identitydomain.User{
		Name:  email,
		Email: email,
	}
	userID, err := u.Repo.Create(ctx, newUser)
	if err != nil {
		return 0, err
	}

	authProvider := &identitydomain.UserAuthProvider{
		UserID:   userID,
		Provider: "google",
	}
	if err := u.AuthProviderRepo.Create(ctx, authProvider); err != nil {
		return 0, err
	}

	return userID, nil
}

func (u *UserService) IssueTokens(ctx context.Context, userID int, issuedBy string) (*IssuedTokens, error) {
	user, err := u.Repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	accessToken, expiresAt, err := jwtadapter.GenerateJWT(userID, user.Name, 15*time.Minute)
	if err != nil {
		return nil, err
	}

	rawRefreshToken, hashed, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	tokenEntity := &identitydomain.RefreshToken{
		UserID:     userID,
		TokenHash:  hashed,
		ExpiresAt:  time.Now().Add(30 * 24 * time.Hour),
		Revoked:    false,
		CreateUser: issuedBy,
		UpdateUser: issuedBy,
	}

	if err := u.RefreshRepo.Create(ctx, tokenEntity); err != nil {
		return nil, err
	}

	return &IssuedTokens{
		AccessToken:          accessToken,
		RefreshToken:         rawRefreshToken,
		AccessTokenExpiresAt: expiresAt,
	}, nil
}

func (u *UserService) RefreshTokens(ctx context.Context, refreshToken string) (*IssuedTokens, error) {
	hashed := hashToken(refreshToken)

	entity, err := u.RefreshRepo.FindByHash(ctx, hashed)
	if err != nil {
		return nil, ErrAuthFailed
	}
	if entity.Revoked || time.Now().After(entity.ExpiresAt) {
		return nil, ErrAuthFailed
	}

	if err := u.RefreshRepo.Revoke(ctx, entity.ID, "auth"); err != nil {
		return nil, err
	}

	return u.IssueTokens(ctx, entity.UserID, "auth")
}

func generateRefreshToken() (string, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	return token, hashToken(token), nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
