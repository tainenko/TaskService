package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "0123456789abcdef0123456789abcdef"

func TestNewTokenManager_Validation(t *testing.T) {
	_, err := NewTokenManager("short", time.Hour)
	assert.Error(t, err)
	_, err = NewTokenManager(testSecret, 0)
	assert.Error(t, err)
}

func TestTokenManager_RoundTrip(t *testing.T) {
	m, err := NewTokenManager(testSecret, time.Hour)
	require.NoError(t, err)

	tok, exp, err := m.Generate(42)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now().Add(time.Hour), exp, 5*time.Second)

	id, err := m.Parse(tok)
	require.NoError(t, err)
	assert.Equal(t, int32(42), id)
}

func TestTokenManager_Rejects(t *testing.T) {
	m, _ := NewTokenManager(testSecret, time.Hour)
	other, _ := NewTokenManager(strings.Repeat("x", 32), time.Hour)

	valid, _, _ := m.Generate(1)
	foreign, _, _ := other.Generate(1)
	expired, _, _ := (&TokenManager{secret: []byte(testSecret), ttl: -time.Minute}).Generate(1)
	none, _ := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.RegisteredClaims{
		Subject: "1", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	noExp, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Subject: "1"}).SignedString([]byte(testSecret))
	badSub, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject: "abc", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString([]byte(testSecret))

	for name, tok := range map[string]string{
		"garbage": "not-a-token", "tampered": valid + "x", "wrong secret": foreign,
		"expired": expired, "alg none": none, "no expiry": noExp, "bad subject": badSub,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := m.Parse(tok)
			assert.ErrorIs(t, err, ErrInvalidToken)
		})
	}
}
