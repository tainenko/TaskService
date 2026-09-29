//go:build integration

package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// session registers a user and returns their credentials and initial tokens.
func session(t *testing.T) (email string, tokens authData) {
	t.Helper()
	c := &client{t: t}
	email = uniqueEmail(t)
	status, env := c.do(http.MethodPost, "/auth/register", map[string]string{"email": email, "password": "password123"})
	require.Equal(t, http.StatusCreated, status, env.Msg)
	tokens = decode[authData](t, env)
	require.NotEmpty(t, tokens.RefreshToken)
	return email, tokens
}

func refresh(t *testing.T, token string) (int, authData) {
	t.Helper()
	status, env := (&client{t: t}).do(http.MethodPost, "/auth/refresh", map[string]string{"refresh_token": token})
	if status != http.StatusOK {
		return status, authData{}
	}
	return status, decode[authData](t, env)
}

func TestRefreshRotatesTokens(t *testing.T) {
	_, first := session(t)

	status, second := refresh(t, first.RefreshToken)
	require.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, second.Token)
	assert.NotEqual(t, first.RefreshToken, second.RefreshToken, "refresh token is rotated")

	// The new access token works on protected routes.
	c := &client{t: t, token: second.Token}
	status, _ = c.do(http.MethodGet, "/tasks", nil)
	assert.Equal(t, http.StatusOK, status)

	// The chain continues.
	status, third := refresh(t, second.RefreshToken)
	require.Equal(t, http.StatusOK, status)
	assert.NotEqual(t, second.RefreshToken, third.RefreshToken)

	// Only a hash is stored, never the token itself.
	var n int
	require.NoError(t, rawDB.QueryRow(`SELECT COUNT(*) FROM refresh_token WHERE token_hash = $1`, third.RefreshToken).Scan(&n))
	assert.Zero(t, n)
}

func TestRefreshTokenReuseRevokesFamily(t *testing.T) {
	_, first := session(t)

	status, second := refresh(t, first.RefreshToken)
	require.Equal(t, http.StatusOK, status)

	// Replaying the already-used token is treated as theft...
	status, _ = refresh(t, first.RefreshToken)
	assert.Equal(t, http.StatusUnauthorized, status)

	// ...and kills the legitimate descendant too, forcing a fresh login.
	status, _ = refresh(t, second.RefreshToken)
	assert.Equal(t, http.StatusUnauthorized, status)
}

func TestRefreshFamiliesAreIndependent(t *testing.T) {
	email, first := session(t)

	// A second login (another device) starts its own family.
	c := &client{t: t}
	status, env := c.do(http.MethodPost, "/auth/login", map[string]string{"email": email, "password": "password123"})
	require.Equal(t, http.StatusOK, status)
	other := decode[authData](t, env)

	// Reuse on the first device does not log out the second.
	refresh(t, first.RefreshToken)
	status, _ = refresh(t, first.RefreshToken)
	assert.Equal(t, http.StatusUnauthorized, status)
	status, _ = refresh(t, other.RefreshToken)
	assert.Equal(t, http.StatusOK, status)
}

func TestRefreshRejectsBadTokens(t *testing.T) {
	c := &client{t: t}
	status, _ := refresh(t, "not-a-real-token")
	assert.Equal(t, http.StatusUnauthorized, status)

	status, _ = c.do(http.MethodPost, "/auth/refresh", map[string]string{})
	assert.Equal(t, http.StatusBadRequest, status)

	// An access token is not a refresh token.
	_, tokens := session(t)
	status, _ = refresh(t, tokens.Token)
	assert.Equal(t, http.StatusUnauthorized, status)
}

func TestRefreshTokenExpiry(t *testing.T) {
	_, tokens := session(t)

	_, err := rawDB.Exec(`UPDATE refresh_token SET expires_at = NOW() - INTERVAL '1 minute'
		WHERE token_hash = encode(sha256($1::bytea), 'hex')`, tokens.RefreshToken)
	require.NoError(t, err)

	status, _ := refresh(t, tokens.RefreshToken)
	assert.Equal(t, http.StatusUnauthorized, status)
}

func TestLogout(t *testing.T) {
	email, tokens := session(t)
	c := &client{t: t}

	status, env := c.do(http.MethodPost, "/auth/logout", map[string]string{"refresh_token": tokens.RefreshToken})
	require.Equal(t, http.StatusOK, status, env.Msg)

	status, _ = refresh(t, tokens.RefreshToken)
	assert.Equal(t, http.StatusUnauthorized, status, "refresh token is dead after logout")

	// Idempotent, and unknown tokens do not error or leak anything.
	status, _ = c.do(http.MethodPost, "/auth/logout", map[string]string{"refresh_token": tokens.RefreshToken})
	assert.Equal(t, http.StatusOK, status)
	status, _ = c.do(http.MethodPost, "/auth/logout", map[string]string{"refresh_token": "unknown"})
	assert.Equal(t, http.StatusOK, status)
	status, _ = c.do(http.MethodPost, "/auth/logout", map[string]string{})
	assert.Equal(t, http.StatusBadRequest, status)

	// Logging in again works.
	status, _ = c.do(http.MethodPost, "/auth/login", map[string]string{"email": email, "password": "password123"})
	assert.Equal(t, http.StatusOK, status)
}

func TestLogoutOnlyEndsOwnSession(t *testing.T) {
	email, phone := session(t)
	c := &client{t: t}
	status, env := c.do(http.MethodPost, "/auth/login", map[string]string{"email": email, "password": "password123"})
	require.Equal(t, http.StatusOK, status)
	laptop := decode[authData](t, env)

	status, _ = c.do(http.MethodPost, "/auth/logout", map[string]string{"refresh_token": phone.RefreshToken})
	require.Equal(t, http.StatusOK, status)

	status, _ = refresh(t, phone.RefreshToken)
	assert.Equal(t, http.StatusUnauthorized, status)
	status, _ = refresh(t, laptop.RefreshToken)
	assert.Equal(t, http.StatusOK, status, "the other device stays logged in")
}

func TestLogoutAll(t *testing.T) {
	email, phone := session(t)
	anon := &client{t: t}
	status, env := anon.do(http.MethodPost, "/auth/login", map[string]string{"email": email, "password": "password123"})
	require.Equal(t, http.StatusOK, status)
	laptop := decode[authData](t, env)

	status, _ = anon.do(http.MethodPost, "/auth/logout-all", nil)
	assert.Equal(t, http.StatusUnauthorized, status, "requires an access token")

	authed := &client{t: t, token: phone.Token}
	status, _ = authed.do(http.MethodPost, "/auth/logout-all", nil)
	require.Equal(t, http.StatusOK, status)

	for _, tok := range []string{phone.RefreshToken, laptop.RefreshToken} {
		status, _ = refresh(t, tok)
		assert.Equal(t, http.StatusUnauthorized, status)
	}

	// Other users are unaffected.
	_, bystander := session(t)
	status, _ = refresh(t, bystander.RefreshToken)
	assert.Equal(t, http.StatusOK, status)
}

func TestConcurrentRefreshOnlyOneWins(t *testing.T) {
	_, tokens := session(t)

	const n = 5
	results := make(chan int, n)
	start := make(chan struct{})
	for i := 0; i < n; i++ {
		go func() {
			<-start
			status, _ := refresh(t, tokens.RefreshToken)
			results <- status
		}()
	}
	close(start)

	ok := 0
	for i := 0; i < n; i++ {
		select {
		case s := <-results:
			if s == http.StatusOK {
				ok++
			}
		case <-time.After(10 * time.Second):
			t.Fatal("timeout")
		}
	}
	assert.LessOrEqual(t, ok, 1, "a refresh token can be exchanged at most once")
}
