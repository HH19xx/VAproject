package usecases

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"

	"server/main/domain"
	"server/main/infra/jwt"

	"golang.org/x/crypto/bcrypt"
)

// 認証失敗時のエラー
var ErrAuthFailed = errors.New("認証失敗")

// ユーザーに関する永続化の抽象インターフェース
type UserRepository interface {
	FindByName(ctx context.Context, name string) (*domain.User, error)
	FindByID(ctx context.Context, userID int) (*domain.User, error)
	UpdateName(ctx context.Context, userID int, newName string) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) (int, error)
}

// ユーザー認証プロバイダーの永続化の抽象インターフェース
type UserAuthProviderRepository interface {
	FindByUserID(ctx context.Context, userID int) ([]*domain.UserAuthProvider, error)
	HasProvider(ctx context.Context, userID int, provider string) (bool, error)
	Create(ctx context.Context, authProvider *domain.UserAuthProvider) error
	Delete(ctx context.Context, userID int, provider string) error
}

// リフレッシュトークンのIOを定義する
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
	Revoke(ctx context.Context, id int, updateUser string) error
}

// ユーザーに関するユースケースを提供する
type UserUsecase struct {
	Repo             UserRepository
	AuthProviderRepo UserAuthProviderRepository
	RefreshRepo      RefreshTokenRepository
}

// アクセストークンとリフレッシュトークンのセットを表す
type IssuedTokens struct {
	AccessToken          string
	RefreshToken         string
	AccessTokenExpiresAt time.Time
}

// プロフィール取得時に返すユーザー情報
type UserProfile struct {
	User          *domain.User
	AuthProviders []string // 'local', 'google' 等のリスト
}

// ユーザー名とパスワードを用いて認証を行い、成功すればユーザーIDを返す
func (u *UserUsecase) Login(ctx context.Context, name, password string) (int, error) {
	user, err := u.Repo.FindByName(ctx, name)
	if err != nil {
		return 0, ErrAuthFailed
	}

	// bcryptでハッシュ化されたパスワードと入力パスワードを比較する
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return 0, ErrAuthFailed
	}

	return user.ID, nil
}

// 新規ユーザーを登録し、登録されたユーザーIDを返す
func (u *UserUsecase) Register(ctx context.Context, name, password string) (int, error) {
	// ユーザー名の重複チェック（既存ユーザーが見つかったら登録不可）
	existingUser, err := u.Repo.FindByName(ctx, name)
	if err == nil && existingUser != nil {
		return 0, errors.New("ユーザー名が既に使用されています")
	}

	// パスワードをbcryptでハッシュ化する
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	// 新規ユーザーエンティティを組み立てる
	newUser := &domain.User{
		Name:     name,
		Password: string(hashedPassword),
	}

	// データベースにユーザーを保存し、生成されたIDを取得する
	userID, err := u.Repo.Create(ctx, newUser)
	if err != nil {
		return 0, err
	}

	// local認証プロバイダーを追加する
	authProvider := &domain.UserAuthProvider{
		UserID:   userID,
		Provider: "local",
	}
	if err := u.AuthProviderRepo.Create(ctx, authProvider); err != nil {
		return 0, err
	}

	return userID, nil
}

// 指定ユーザーIDのプロフィール情報を取得する
func (u *UserUsecase) GetProfile(ctx context.Context, userID int) (*UserProfile, error) {
	// リポジトリからユーザー情報を取得する
	user, err := u.Repo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.New("ユーザーが見つかりません")
	}

	// パスワードフィールドを空にする（セキュリティ対策）
	user.Password = ""

	// 認証プロバイダー情報を取得する
	authProviders, err := u.AuthProviderRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("認証プロバイダー情報の取得に失敗しました")
	}

	// プロバイダー名のリストに変換する
	providers := make([]string, len(authProviders))
	for i, ap := range authProviders {
		providers[i] = ap.Provider
	}

	return &UserProfile{
		User:          user,
		AuthProviders: providers,
	}, nil
}

// 指定ユーザーIDのプロフィール情報を更新する
func (u *UserUsecase) UpdateProfile(ctx context.Context, userID int, newName string) error {
	// 新しいユーザー名が空でないかチェックする
	if newName == "" {
		return errors.New("ユーザー名を入力してください")
	}

	// 変更先のユーザー名が既に使用されているかチェックする
	existingUser, err := u.Repo.FindByName(ctx, newName)
	if err == nil && existingUser != nil && existingUser.ID != userID {
		return errors.New("そのユーザー名は既に使用されています")
	}

	// ユーザー名を更新する
	if err := u.Repo.UpdateName(ctx, userID, newName); err != nil {
		return errors.New("ユーザー名の更新に失敗しました")
	}

	return nil
}

// OAuthで取得したメールアドレスからユーザーを照会し、なければ作成する
func (u *UserUsecase) FindOrCreateUserByEmail(ctx context.Context, email string) (int, error) {
	// 既存ユーザーを検索する
	user, err := u.Repo.FindByEmail(ctx, email)
	if err == nil {
		// 既存ユーザーが見つかった場合、google認証プロバイダーを追加する（未登録の場合）
		hasGoogle, _ := u.AuthProviderRepo.HasProvider(ctx, user.ID, "google")
		if !hasGoogle {
			authProvider := &domain.UserAuthProvider{
				UserID:   user.ID,
				Provider: "google",
			}
			u.AuthProviderRepo.Create(ctx, authProvider)
		}
		return user.ID, nil
	}

	// 新規ユーザーを作成する
	newUser := &domain.User{
		Name:  email, // 暫定的にnameにもemailを設定
		Email: email,
	}
	userID, err := u.Repo.Create(ctx, newUser)
	if err != nil {
		return 0, err
	}

	// google認証プロバイダーを追加する
	authProvider := &domain.UserAuthProvider{
		UserID:   userID,
		Provider: "google",
	}
	if err := u.AuthProviderRepo.Create(ctx, authProvider); err != nil {
		return 0, err
	}

	return userID, nil
}

// アクセストークンとリフレッシュトークンを指定ユーザー向けに発行する
func (u *UserUsecase) IssueTokens(ctx context.Context, userID int, issuedBy string) (*IssuedTokens, error) {
	// アクセストークンを15分有効で発行する
	accessToken, expiresAt, err := jwt.GenerateJWT(userID, 15*time.Minute)
	if err != nil {
		return nil, err
	}

	// リフレッシュトークン用の乱数を生成し、Base64で文字列化する
	rawRefreshToken, hashed, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	// 永続化用エンティティを組み立てる
	tokenEntity := &domain.RefreshToken{
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

// リフレッシュトークンを検証し、新しいトークンを発行する
func (u *UserUsecase) RefreshTokens(ctx context.Context, refreshToken string) (*IssuedTokens, error) {
	hashed := hashToken(refreshToken)

	entity, err := u.RefreshRepo.FindByHash(ctx, hashed)
	if err != nil {
		return nil, ErrAuthFailed
	}

	if entity.Revoked {
		return nil, ErrAuthFailed
	}

	if time.Now().After(entity.ExpiresAt) {
		return nil, ErrAuthFailed
	}

	// 使用済みリフレッシュトークンは即座に失効させる
	if err := u.RefreshRepo.Revoke(ctx, entity.ID, "auth"); err != nil {
		return nil, err
	}

	// 新しいトークンを発行し直します。
	return u.IssueTokens(ctx, entity.UserID, "auth")
}

// 乱数でリフレッシュトークンを生成し、ハッシュも返す
func generateRefreshToken() (string, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	return token, hashToken(token), nil
}

// リフレッシュトークン文字列をSHA-256でハッシュ化する
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
