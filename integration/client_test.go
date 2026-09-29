//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// envelope mirrors response.Result.
type envelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type client struct {
	t     *testing.T
	token string
}

// do sends a request and returns the status code and decoded envelope.
func (c *client) do(method, path string, body any) (int, envelope) {
	c.t.Helper()
	var buf io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(c.t, err)
		buf = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, server.URL+path, buf)
	require.NoError(c.t, err)
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(c.t, err)
	defer func() { _ = resp.Body.Close() }()

	var env envelope
	raw, err := io.ReadAll(resp.Body)
	require.NoError(c.t, err)
	if len(raw) > 0 {
		require.NoError(c.t, json.Unmarshal(raw, &env), string(raw))
	}
	return resp.StatusCode, env
}

func decode[T any](t *testing.T, env envelope) T {
	t.Helper()
	var v T
	require.NoError(t, json.Unmarshal(env.Data, &v))
	return v
}

type authData struct {
	Token string `json:"token"`
	User  struct {
		ID    int32  `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
}

type task struct {
	ID          int32   `json:"id"`
	Name        string  `json:"name"`
	Status      int32   `json:"status"`
	Description string  `json:"description"`
	DueDate     *string `json:"due_date"`
	Priority    int32   `json:"priority"`
}

type taskList struct {
	Page struct {
		TotalRecords int64 `json:"total_records"`
		TotalPages   int   `json:"total_pages"`
		NextPage     *int  `json:"next_page"`
		PrevPage     *int  `json:"prev_page"`
	} `json:"page"`
	Tasks []task `json:"tasks"`
}

// newUser registers a fresh user and returns a client authenticated as them.
func newUser(t *testing.T) *client {
	t.Helper()
	c := &client{t: t}
	status, env := c.do(http.MethodPost, "/auth/register", map[string]string{
		"email": uniqueEmail(t), "password": "password123",
	})
	require.Equal(t, http.StatusCreated, status, env.Msg)
	c.token = decode[authData](t, env).Token
	return c
}

func (c *client) createTask(body map[string]any) int32 {
	c.t.Helper()
	status, env := c.do(http.MethodPost, "/tasks", body)
	require.Equal(c.t, http.StatusCreated, status, env.Msg)
	return decode[struct {
		ID int32 `json:"id"`
	}](c.t, env).ID
}
