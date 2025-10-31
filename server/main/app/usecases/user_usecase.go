package usecases

import (
    "context"
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
    "errors"
    "time"

    "golang.org/x/crypto/bcrypt"
    "server/main/domain"
    "server/main/infra/jwt"
)

// 認証失敗時のエラー
var ErrAuthFailed = errors.New("認証失敗")

// UserRepository はユーザーに関する永続化の抽象インターフェースです。
type UserRepository interface {
    FindByName(ctx context.Context, name string) (*domain.User, error)

    // OAuth用：メールアドレスからユーザーを探す／作成する
    FindByEmail(ctx context.Context, email string) (*domain.User, error)
    Create(ctx context.Context, user *domain.User) (int, error)
}

// RefreshTokenRepository はリフレッシュトークンのIOを定義します。
type RefreshTokenRepository interface {
    Create(ctx context.Context, token *domain.RefreshToken) error
    FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
    Revoke(ctx context.Context, id int, updateUser string) error
}

// UserUsecase はユーザーに関するユースケースを提供します。
type UserUsecase struct {
    Repo        UserRepository
    RefreshRepo RefreshTokenRepository
}

// IssuedTokens はアクセストークンとリフレッシュトークンのセットを表します。
type IssuedTokens struct {
    AccessToken          string
    RefreshToken         string
    AccessTokenExpiresAt time.Time
}

// Login はユーザー名とパスワードを用いて認証を行い、成功すればユーザーIDを返します。
func (u *UserUsecase) Login(ctx context.Context, name, password string) (int, error) {
    user, err := u.Repo.FindByName(ctx, name)
    if err != nil {
		return 0, ErrAuthFailed
	}

	// bcryptでハッシュ化されたパスワードと入力パスワードを比較します。
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return 0, ErrAuthFailed
	}

	return user.ID, nil
}

// FindOrCreateUserByEmail はOAuthで取得したメールアドレスからユーザーを照会し、なければ作成します。
func (u *UserUsecase) FindOrCreateUserByEmail(ctx context.Context, email string) (int, error) {
    user, err := u.Repo.FindByEmail(ctx, email)
    if err == nil {
        return user.ID, nil
    }

    newUser := &domain.User{
        Name:  email, // 暫定的にnameにもemailを設定
        Email: email,
    }
    return u.Repo.Create(ctx, newUser)
}

// IssueTokens は指定ユーザー向けにアクセストークンとリフレッシュトークンを発行します。
func (u *UserUsecase) IssueTokens(ctx context.Context, userID int, issuedBy string) (*IssuedTokens, error) {
    // アクセストークンを15分有効で発行します。
    accessToken, expiresAt, err := jwt.GenerateJWT(userID, 15*time.Minute)
    if err != nil {
        return nil, err
    }

    // リフレッシュトークン用の乱数を生成し、Base64で文字列化します。
    rawRefreshToken, hashed, err := generateRefreshToken()
    if err != nil {
        return nil, err
    }

    // 永続化用エンティティを組み立てます。
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

// RefreshTokens はリフレッシュトークンを検証し、新しいトークンを発行します。
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

    // 使用済みリフレッシュトークンは即座に失効させます。
    if err := u.RefreshRepo.Revoke(ctx, entity.ID, "auth"); err != nil {
        return nil, err
    }

    // 新しいトークンを発行し直します。
    return u.IssueTokens(ctx, entity.UserID, "auth")
}

// generateRefreshToken は乱数でリフレッシュトークンを生成し、ハッシュも返します。
func generateRefreshToken() (string, string, error) {
    raw := make([]byte, 32)
    if _, err := rand.Read(raw); err != nil {
        return "", "", err
    }
    token := base64.RawURLEncoding.EncodeToString(raw)
    return token, hashToken(token), nil
}

// hashToken はリフレッシュトークン文字列をSHA-256でハッシュ化します。
func hashToken(token string) string {
    h := sha256.Sum256([]byte(token))
    return base64.RawURLEncoding.EncodeToString(h[:])
}
