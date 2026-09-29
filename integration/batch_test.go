//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type batchResult struct {
	Processed []int32 `json:"processed"`
	NotFound  []int32 `json:"not_found"`
}

func TestBatchCreate(t *testing.T) {
	c := newUser(t)

	status, env := c.do(http.MethodPost, "/tasks/batch", map[string]any{"tasks": []map[string]any{
		{"name": "one", "status": 0, "tags": []string{"Batch", "a"}},
		{"name": "two", "status": 1, "priority": 2},
		{"name": "three", "status": 2, "description": "d"},
	}})
	require.Equal(t, http.StatusCreated, status, env.Msg)
	ids := decode[struct {
		IDs []int32 `json:"ids"`
	}](t, env).IDs
	require.Len(t, ids, 3)

	// IDs follow the input order.
	assert.Equal(t, "one", c.getTask(ids[0]).Name)
	assert.Equal(t, []string{"a", "batch"}, c.getTask(ids[0]).Tags)
	assert.EqualValues(t, 2, c.getTask(ids[1]).Priority)
	assert.Equal(t, "three", c.getTask(ids[2]).Name)
	assert.EqualValues(t, 2, c.getTask(ids[2]).Status)
}

func TestBatchCreateIsAtomic(t *testing.T) {
	c := newUser(t)

	// The third item is invalid, so nothing may be stored.
	status, _ := c.do(http.MethodPost, "/tasks/batch", map[string]any{"tasks": []map[string]any{
		{"name": "ok1", "status": 0, "tags": []string{"atomic"}},
		{"name": "ok2", "status": 0},
		{"name": "", "status": 0},
	}})
	assert.Equal(t, http.StatusBadRequest, status)

	status, env := c.do(http.MethodGet, "/tasks", nil)
	require.Equal(t, http.StatusOK, status)
	assert.EqualValues(t, 0, decode[taskList](t, env).Page.TotalRecords)
	assert.Empty(t, c.tags(), "tags of a rejected batch are not created either")
}

func TestBatchCreateLimits(t *testing.T) {
	c := newUser(t)
	mk := func(n int) []map[string]any {
		items := make([]map[string]any, n)
		for i := range items {
			items[i] = map[string]any{"name": fmt.Sprintf("t%d", i), "status": 0}
		}
		return items
	}

	status, _ := c.do(http.MethodPost, "/tasks/batch", map[string]any{"tasks": mk(101)})
	assert.Equal(t, http.StatusBadRequest, status)
	status, _ = c.do(http.MethodPost, "/tasks/batch", map[string]any{"tasks": mk(0)})
	assert.Equal(t, http.StatusBadRequest, status)
	status, _ = c.do(http.MethodPost, "/tasks/batch", map[string]any{"tasks": mk(100)})
	assert.Equal(t, http.StatusCreated, status)
}

func TestBatchStatusAndDelete(t *testing.T) {
	alice, bob := newUser(t), newUser(t)
	a1 := alice.createTask(map[string]any{"name": "a1", "status": 0})
	a2 := alice.createTask(map[string]any{"name": "a2", "status": 0})
	a3 := alice.createTask(map[string]any{"name": "a3", "status": 0})
	b1 := bob.createTask(map[string]any{"name": "b1", "status": 0})
	const missing = int32(999999)

	// Status: alice's ids are updated; bob's task and a missing id are reported as not found.
	status, env := alice.do(http.MethodPatch, "/tasks/batch/status", map[string]any{
		"ids": []int32{a1, a2, b1, missing, a1}, "status": 1,
	})
	require.Equal(t, http.StatusOK, status, env.Msg)
	res := decode[batchResult](t, env)
	assert.Equal(t, []int32{a1, a2}, res.Processed, "duplicates collapsed, request order kept")
	assert.Equal(t, []int32{b1, missing}, res.NotFound)
	assert.EqualValues(t, 1, alice.getTask(a1).Status)
	assert.EqualValues(t, 1, alice.getTask(a2).Status)
	assert.EqualValues(t, 0, alice.getTask(a3).Status, "task not in the batch is untouched")
	assert.EqualValues(t, 0, bob.getTask(b1).Status, "another user's task is untouched")

	// Status 0 is a valid target.
	status, _ = alice.do(http.MethodPatch, "/tasks/batch/status", map[string]any{"ids": []int32{a1}, "status": 0})
	require.Equal(t, http.StatusOK, status)
	assert.EqualValues(t, 0, alice.getTask(a1).Status)

	// Delete.
	status, env = alice.do(http.MethodDelete, "/tasks/batch", map[string]any{"ids": []int32{a1, a2, b1, missing}})
	require.Equal(t, http.StatusOK, status, env.Msg)
	res = decode[batchResult](t, env)
	assert.Equal(t, []int32{a1, a2}, res.Processed)
	assert.Equal(t, []int32{b1, missing}, res.NotFound)

	for _, id := range []int32{a1, a2} {
		status, _ = alice.do(http.MethodGet, fmt.Sprintf("/tasks/%d", id), nil)
		assert.Equal(t, http.StatusNotFound, status)
	}
	assert.Equal(t, "a3", alice.getTask(a3).Name)
	assert.Equal(t, "b1", bob.getTask(b1).Name, "bob's task survives alice's batch delete")

	// Deleting again reports everything as not found rather than failing.
	status, env = alice.do(http.MethodDelete, "/tasks/batch", map[string]any{"ids": []int32{a1, a2}})
	require.Equal(t, http.StatusOK, status)
	res = decode[batchResult](t, env)
	assert.Empty(t, res.Processed)
	assert.Equal(t, []int32{a1, a2}, res.NotFound)
}

func TestBatchValidationAndAuth(t *testing.T) {
	c := newUser(t)
	for name, tc := range map[string]struct {
		method, path string
		body         map[string]any
	}{
		"delete empty":     {http.MethodDelete, "/tasks/batch", map[string]any{"ids": []int32{}}},
		"delete zero id":   {http.MethodDelete, "/tasks/batch", map[string]any{"ids": []int32{0}}},
		"delete no body":   {http.MethodDelete, "/tasks/batch", nil},
		"status missing":   {http.MethodPatch, "/tasks/batch/status", map[string]any{"ids": []int32{1}}},
		"status invalid":   {http.MethodPatch, "/tasks/batch/status", map[string]any{"ids": []int32{1}, "status": 7}},
		"status empty ids": {http.MethodPatch, "/tasks/batch/status", map[string]any{"ids": []int32{}, "status": 1}},
	} {
		status, _ := c.do(tc.method, tc.path, tc.body)
		assert.Equal(t, http.StatusBadRequest, status, name)
	}

	anon := &client{t: t}
	for _, r := range [][2]string{{http.MethodPost, "/tasks/batch"}, {http.MethodDelete, "/tasks/batch"}, {http.MethodPatch, "/tasks/batch/status"}} {
		status, _ := anon.do(r[0], r[1], nil)
		assert.Equal(t, http.StatusUnauthorized, status, r[1])
	}

	// The batch routes must not shadow the single-task routes.
	id := c.createTask(map[string]any{"name": "single", "status": 0})
	assert.Equal(t, "single", c.getTask(id).Name)
	status, _ := c.do(http.MethodPatch, fmt.Sprintf("/tasks/%d/status", id), map[string]any{"status": 1})
	assert.Equal(t, http.StatusOK, status)
}
