//go:build integration

package integration

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github/TaskService/ratelimit"
)

func TestLoginLockoutPerAccount(t *testing.T) {
	c := &client{t: t}
	email := uniqueEmail(t)
	good := map[string]string{"email": email, "password": "password123"}
	bad := map[string]string{"email": email, "password": "wrong-password"}

	status, _ := c.do(http.MethodPost, "/auth/register", good)
	require.Equal(t, http.StatusCreated, status)

	// The shared server allows 5 failures per account.
	for i := 0; i < 5; i++ {
		status, _ = c.do(http.MethodPost, "/auth/login", bad)
		require.Equal(t, http.StatusUnauthorized, status, "failure %d", i+1)
	}

	// Locked out: even the correct password is refused, with a Retry-After hint.
	status, env := c.do(http.MethodPost, "/auth/login", good)
	assert.Equal(t, http.StatusTooManyRequests, status, env.Msg)

	// Case/whitespace variants of the email hit the same bucket.
	status, _ = c.do(http.MethodPost, "/auth/login", map[string]string{"email": strings.ToUpper(email), "password": "password123"})
	assert.Equal(t, http.StatusTooManyRequests, status)

	// Other accounts are unaffected.
	other := newUser(t)
	assert.NotEmpty(t, other.token)
}

func TestLoginSuccessResetsFailures(t *testing.T) {
	c := &client{t: t}
	email := uniqueEmail(t)
	good := map[string]string{"email": email, "password": "password123"}
	bad := map[string]string{"email": email, "password": "wrong-password"}
	status, _ := c.do(http.MethodPost, "/auth/register", good)
	require.Equal(t, http.StatusCreated, status)

	// Four failures then a success must not accumulate towards a lockout.
	for round := 0; round < 3; round++ {
		for i := 0; i < 4; i++ {
			status, _ = c.do(http.MethodPost, "/auth/login", bad)
			require.Equal(t, http.StatusUnauthorized, status)
		}
		status, _ = c.do(http.MethodPost, "/auth/login", good)
		require.Equal(t, http.StatusOK, status, "round %d", round)
	}
}

func TestAuthEndpointsRateLimitedPerIP(t *testing.T) {
	srv := httptest.NewServer(newEngine(
		ratelimit.NewKeyedLimiter(ratelimit.PerMinute(1), 3, time.Hour),
		ratelimit.NewKeyedLimiter(ratelimit.PerMinute(1), 100, time.Hour),
	))
	defer srv.Close()

	post := func(path string) *http.Response {
		req, err := http.NewRequest(http.MethodPost, srv.URL+path, nil)
		require.NoError(t, err)
		// Spoofed header must be ignored because no proxy is trusted.
		req.Header.Set("X-Forwarded-For", "203.0.113."+time.Now().Format("05.000")[3:])
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		_ = resp.Body.Close()
		return resp
	}

	// Register and login share the per-IP budget; invalid bodies still count.
	for i := 0; i < 3; i++ {
		assert.Equal(t, http.StatusBadRequest, post("/auth/login").StatusCode)
	}
	resp := post("/auth/register")
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	assert.NotEmpty(t, resp.Header.Get("Retry-After"))

	// Task routes are not covered by the auth limiter.
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/tasks", nil)
	r2, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = r2.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, r2.StatusCode)
}
