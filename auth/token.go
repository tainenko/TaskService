// Package auth issues and validates JWT access tokens.
package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// MinSecretLen is the minimum accepted length of the HMAC signing secret.
const MinSecretLen = 32

var ErrInvalidToken = errors.New("invalid token")

type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenManager(secret string, ttl time.Duration) (*TokenManager, error) {
	if len(secret) < MinSecretLen {
		return nil, fmt.Errorf("jwt secret must be at least %d characters", MinSecretLen)
	}
	if ttl <= 0 {
		return nil, errors.New("jwt ttl must be positive")
	}
	return &TokenManager{secret: []byte(secret), ttl: ttl}, nil
}

// Generate returns a signed token for the user and its expiry time.
func (m *TokenManager) Generate(userID int32) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(m.ttl)
	claims := jwt.RegisteredClaims{
		Subject:   strconv.Itoa(int(userID)),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(exp),
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	return signed, exp, err
}

// Parse validates the token and returns the user ID it was issued for.
func (m *TokenManager) Parse(token string) (int32, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(*jwt.Token) (interface{}, error) {
		return m.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
	if err != nil {
		return 0, ErrInvalidToken
	}
	id, err := strconv.ParseInt(claims.Subject, 10, 32)
	if err != nil || id <= 0 {
		return 0, ErrInvalidToken
	}
	return int32(id), nil
}
