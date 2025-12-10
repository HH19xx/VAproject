package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("secret_key")
var ErrInvalidToken = errors.New("invalid token")

// クレーム（トークンに含める情報）
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// JWTトークンを生成
func GenerateJWT(userID int, expiresIn time.Duration) (string, time.Time, error) {
	// 指定された有効期限を現在時刻から加算しRegisteredClaimsに設定
	expirationTime := time.Now().Add(expiresIn)
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(jwtKey)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expirationTime, nil
}

// JWTトークンを検証
func ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
