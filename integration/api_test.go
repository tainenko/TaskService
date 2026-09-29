//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealth(t *testing.T) {
	c := &client{t: t}
	status, _ := c.do(http.MethodGet, "/healthz", nil)
	assert.Equal(t, http.StatusOK, status)
	status, _ = c.do(http.MethodGet, "/readyz", nil)
	assert.Equal(t, http.StatusOK, status)
}

func TestAuthFlow(t *testing.T) {
	c := &client{t: t}
	email := uniqueEmail(t)
	creds := map[string]string{"email": email, "password": "password123"}

	status, env := c.do(http.MethodPost, "/auth/register", creds)
	require.Equal(t, http.StatusCreated, status, env.Msg)
	registered := decode[authData](t, env)
	assert.NotEmpty(t, registered.Token)

	// Password hash must never be exposed.
	assert.NotContains(t, string(env.Data), "password")

	// Emails are case-insensitive, so a different casing is a duplicate.
	status, _ = c.do(http.MethodPost, "/auth/register", map[string]string{"email": strings.ToUpper(email), "password": "password123"})
	assert.Equal(t, http.StatusConflict, status)
	status, _ = c.do(http.MethodPost, "/auth/register", creds)
	assert.Equal(t, http.StatusConflict, status)

	status, env = c.do(http.MethodPost, "/auth/login", creds)
	require.Equal(t, http.StatusOK, status, env.Msg)
	assert.Equal(t, registered.User.ID, decode[authData](t, env).User.ID)

	status, _ = c.do(http.MethodPost, "/auth/login", map[string]string{"email": email, "password": "wrong-password"})
	assert.Equal(t, http.StatusUnauthorized, status)
	status, _ = c.do(http.MethodPost, "/auth/login", map[string]string{"email": "nobody-" + email, "password": "password123"})
	assert.Equal(t, http.StatusUnauthorized, status)

	status, _ = c.do(http.MethodPost, "/auth/register", map[string]string{"email": "bad", "password": "password123"})
	assert.Equal(t, http.StatusBadRequest, status)
}

func TestTasksRequireAuth(t *testing.T) {
	anon := &client{t: t}
	for _, req := range [][2]string{
		{http.MethodGet, "/tasks"}, {http.MethodPost, "/tasks"}, {http.MethodGet, "/tasks/1"},
		{http.MethodPut, "/tasks/1"}, {http.MethodPatch, "/tasks/1/status"}, {http.MethodDelete, "/tasks/1"},
	} {
		status, _ := anon.do(req[0], req[1], nil)
		assert.Equal(t, http.StatusUnauthorized, status, "%s %s", req[0], req[1])
	}

	bad := &client{t: t, token: "not-a-jwt"}
	status, _ := bad.do(http.MethodGet, "/tasks", nil)
	assert.Equal(t, http.StatusUnauthorized, status)
}

func TestTaskLifecycle(t *testing.T) {
	c := newUser(t)

	id := c.createTask(map[string]any{
		"name": "write tests", "status": 0, "description": "integration", "priority": 2,
		"due_date": "2030-01-02T03:04:05Z",
	})
	path := fmt.Sprintf("/tasks/%d", id)

	status, env := c.do(http.MethodGet, path, nil)
	require.Equal(t, http.StatusOK, status)
	got := decode[task](t, env)
	assert.Equal(t, "write tests", got.Name)
	assert.Equal(t, "integration", got.Description)
	assert.EqualValues(t, 2, got.Priority)
	require.NotNil(t, got.DueDate)
	assert.Contains(t, *got.DueDate, "2030-01-02")

	// PUT replaces the fields, including moving status to 1 then back to 0 (zero value).
	status, env = c.do(http.MethodPut, path, map[string]any{"name": "renamed", "status": 1, "priority": 1})
	require.Equal(t, http.StatusOK, status, env.Msg)
	_, env = c.do(http.MethodGet, path, nil)
	got = decode[task](t, env)
	assert.Equal(t, "renamed", got.Name)
	assert.EqualValues(t, 1, got.Status)
	assert.Empty(t, got.Description, "PUT without description clears it")
	assert.Nil(t, got.DueDate, "PUT without due_date clears it")

	status, _ = c.do(http.MethodPut, path, map[string]any{"name": "renamed", "status": 0})
	require.Equal(t, http.StatusOK, status)
	_, env = c.do(http.MethodGet, path, nil)
	assert.EqualValues(t, 0, decode[task](t, env).Status, "status must be settable to 0")

	status, _ = c.do(http.MethodPatch, path+"/status", map[string]any{"status": 2})
	require.Equal(t, http.StatusOK, status)
	_, env = c.do(http.MethodGet, path, nil)
	got = decode[task](t, env)
	assert.EqualValues(t, 2, got.Status)
	assert.Equal(t, "renamed", got.Name, "PATCH must not touch other fields")

	status, _ = c.do(http.MethodPatch, path+"/status", map[string]any{"status": 9})
	assert.Equal(t, http.StatusBadRequest, status)

	status, _ = c.do(http.MethodDelete, path, nil)
	assert.Equal(t, http.StatusOK, status)
	status, _ = c.do(http.MethodGet, path, nil)
	assert.Equal(t, http.StatusNotFound, status)
	status, _ = c.do(http.MethodDelete, path, nil)
	assert.Equal(t, http.StatusNotFound, status, "deleting twice")
	status, _ = c.do(http.MethodPut, path, map[string]any{"name": "x", "status": 0})
	assert.Equal(t, http.StatusNotFound, status, "updating a deleted task")

	// Deletion is soft: the row is still there with deleted_at set.
	var deleted bool
	require.NoError(t, rawDB.QueryRow(`SELECT deleted_at IS NOT NULL FROM task WHERE id = $1`, id).Scan(&deleted))
	assert.True(t, deleted)
}

func TestTaskValidation(t *testing.T) {
	c := newUser(t)
	for name, body := range map[string]map[string]any{
		"missing name":     {"status": 0},
		"empty name":       {"name": "", "status": 0},
		"name too long":    {"name": string(make([]byte, 256)), "status": 0},
		"invalid status":   {"name": "x", "status": 5},
		"invalid priority": {"name": "x", "status": 0, "priority": 7},
	} {
		if name == "name too long" {
			body["name"] = fmt.Sprintf("%0256d", 0)
		}
		status, _ := c.do(http.MethodPost, "/tasks", body)
		assert.Equal(t, http.StatusBadRequest, status, name)
	}

	status, _ := c.do(http.MethodGet, "/tasks/abc", nil)
	assert.Equal(t, http.StatusBadRequest, status)
	status, _ = c.do(http.MethodGet, "/tasks/999999999", nil)
	assert.Equal(t, http.StatusNotFound, status)
}

func TestOwnershipIsolation(t *testing.T) {
	alice, bob := newUser(t), newUser(t)
	id := alice.createTask(map[string]any{"name": "alice secret", "status": 0})
	path := fmt.Sprintf("/tasks/%d", id)

	status, _ := bob.do(http.MethodGet, path, nil)
	assert.Equal(t, http.StatusNotFound, status)
	status, _ = bob.do(http.MethodPut, path, map[string]any{"name": "pwned", "status": 1})
	assert.Equal(t, http.StatusNotFound, status)
	status, _ = bob.do(http.MethodPatch, path+"/status", map[string]any{"status": 1})
	assert.Equal(t, http.StatusNotFound, status)
	status, _ = bob.do(http.MethodDelete, path, nil)
	assert.Equal(t, http.StatusNotFound, status)

	_, env := bob.do(http.MethodGet, "/tasks", nil)
	assert.Empty(t, decode[taskList](t, env).Tasks, "bob must not see alice's tasks in listings")

	// Alice's task is untouched.
	status, env = alice.do(http.MethodGet, path, nil)
	require.Equal(t, http.StatusOK, status)
	got := decode[task](t, env)
	assert.Equal(t, "alice secret", got.Name)
	assert.EqualValues(t, 0, got.Status)
}

func TestListFilterSortPaginate(t *testing.T) {
	c := newUser(t)
	// 12 tasks: names t01..t12, done for even numbers, plus two with LIKE wildcards.
	for i := 1; i <= 12; i++ {
		c.createTask(map[string]any{"name": fmt.Sprintf("t%02d", i), "status": i % 2})
	}
	c.createTask(map[string]any{"name": "100% done", "status": 0})
	c.createTask(map[string]any{"name": "a_b", "status": 0})

	list := func(query string) taskList {
		status, env := c.do(http.MethodGet, "/tasks"+query, nil)
		require.Equal(t, http.StatusOK, status, env.Msg)
		return decode[taskList](t, env)
	}

	all := list("")
	assert.EqualValues(t, 14, all.Page.TotalRecords)
	assert.Len(t, all.Tasks, 10, "default page size")
	assert.Equal(t, 2, all.Page.TotalPages)
	require.NotNil(t, all.Page.NextPage)
	assert.Nil(t, all.Page.PrevPage)
	assert.Equal(t, "a_b", all.Tasks[0].Name, "default order is newest first")

	page2 := list("?page=2")
	assert.Len(t, page2.Tasks, 4)
	assert.Nil(t, page2.Page.NextPage)
	require.NotNil(t, page2.Page.PrevPage)

	// Pages must not overlap or skip.
	seen := map[int32]bool{}
	for _, tk := range append(all.Tasks, page2.Tasks...) {
		assert.False(t, seen[tk.ID], "duplicate id %d across pages", tk.ID)
		seen[tk.ID] = true
	}
	assert.Len(t, seen, 14)

	asc := list("?sort=name&order=asc")
	assert.Equal(t, "100% done", asc.Tasks[0].Name)

	done := list("?status=1")
	assert.EqualValues(t, 6, done.Page.TotalRecords)
	todo := list("?status=0")
	assert.EqualValues(t, 8, todo.Page.TotalRecords)

	// Wildcards in the search are matched literally.
	assert.EqualValues(t, 1, list("?name=%25").Page.TotalRecords, "%% only matches the literal percent")
	assert.EqualValues(t, 1, list("?name=_").Page.TotalRecords, "_ only matches the literal underscore")
	assert.EqualValues(t, 12, list("?name=t").Page.TotalRecords)

	for _, q := range []string{"?page=0", "?pageSize=5", "?pageSize=101", "?status=7"} {
		status, _ := c.do(http.MethodGet, "/tasks"+q, nil)
		assert.Equal(t, http.StatusBadRequest, status, q)
	}
}
