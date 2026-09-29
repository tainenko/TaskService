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

type tagCount struct {
	ID        int32  `json:"id"`
	Name      string `json:"name"`
	TaskCount int64  `json:"task_count"`
}

func (c *client) tags() []tagCount {
	c.t.Helper()
	status, env := c.do(http.MethodGet, "/tags", nil)
	require.Equal(c.t, http.StatusOK, status, env.Msg)
	return decode[[]tagCount](c.t, env)
}

func (c *client) getTask(id int32) task {
	c.t.Helper()
	status, env := c.do(http.MethodGet, fmt.Sprintf("/tasks/%d", id), nil)
	require.Equal(c.t, http.StatusOK, status, env.Msg)
	return decode[task](c.t, env)
}

func TestTaskTags(t *testing.T) {
	c := newUser(t)

	// Tags are normalized (trim, lowercase) and de-duplicated.
	id := c.createTask(map[string]any{"name": "a", "status": 0, "tags": []string{" Work ", "URGENT", "work"}})
	assert.Equal(t, []string{"urgent", "work"}, c.getTask(id).Tags, "sorted by name")

	// Tasks created without tags report an empty list, not null.
	plain := c.createTask(map[string]any{"name": "plain", "status": 0})
	assert.Equal(t, []string{}, c.getTask(plain).Tags)

	// PUT without "tags" leaves them unchanged.
	path := fmt.Sprintf("/tasks/%d", id)
	status, _ := c.do(http.MethodPut, path, map[string]any{"name": "a2", "status": 1})
	require.Equal(t, http.StatusOK, status)
	got := c.getTask(id)
	assert.Equal(t, "a2", got.Name)
	assert.Equal(t, []string{"urgent", "work"}, got.Tags)

	// PUT with tags replaces them.
	status, _ = c.do(http.MethodPut, path, map[string]any{"name": "a2", "status": 1, "tags": []string{"home", "work"}})
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, []string{"home", "work"}, c.getTask(id).Tags)

	// PUT with [] clears them.
	status, _ = c.do(http.MethodPut, path, map[string]any{"name": "a2", "status": 1, "tags": []string{}})
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, []string{}, c.getTask(id).Tags)

	// PATCH status does not touch tags.
	c.do(http.MethodPut, path, map[string]any{"name": "a2", "status": 1, "tags": []string{"keep"}})
	status, _ = c.do(http.MethodPatch, path+"/status", map[string]any{"status": 0})
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, []string{"keep"}, c.getTask(id).Tags)
}

func TestTagFilterAndCounts(t *testing.T) {
	c := newUser(t)
	a := c.createTask(map[string]any{"name": "a", "status": 0, "tags": []string{"work", "urgent"}})
	b := c.createTask(map[string]any{"name": "b", "status": 0, "tags": []string{"work"}})
	c.createTask(map[string]any{"name": "c", "status": 0, "tags": []string{"home"}})
	c.createTask(map[string]any{"name": "untagged", "status": 0})

	list := func(q string) taskList {
		status, env := c.do(http.MethodGet, "/tasks"+q, nil)
		require.Equal(t, http.StatusOK, status, env.Msg)
		return decode[taskList](t, env)
	}

	work := list("?tag=work")
	assert.EqualValues(t, 2, work.Page.TotalRecords)
	ids := []int32{work.Tasks[0].ID, work.Tasks[1].ID}
	assert.ElementsMatch(t, []int32{a, b}, ids)
	for _, tk := range work.Tasks {
		assert.Contains(t, tk.Tags, "work")
	}
	assert.EqualValues(t, 1, list("?tag=URGENT").Page.TotalRecords, "filter is case-insensitive")
	assert.EqualValues(t, 0, list("?tag=nope").Page.TotalRecords)
	assert.EqualValues(t, 1, list("?tag=work&name=a").Page.TotalRecords, "combines with other filters")
	assert.EqualValues(t, 4, list("").Page.TotalRecords)

	counts := map[string]int64{}
	for _, tg := range c.tags() {
		counts[tg.Name] = tg.TaskCount
	}
	assert.Equal(t, map[string]int64{"work": 2, "urgent": 1, "home": 1}, counts)

	// Deleted tasks are not counted.
	status, _ := c.do(http.MethodDelete, fmt.Sprintf("/tasks/%d", b), nil)
	require.Equal(t, http.StatusOK, status)
	counts = map[string]int64{}
	for _, tg := range c.tags() {
		counts[tg.Name] = tg.TaskCount
	}
	assert.EqualValues(t, 1, counts["work"])
	assert.EqualValues(t, 1, list("?tag=work").Page.TotalRecords)
}

func TestDeleteTag(t *testing.T) {
	c := newUser(t)
	id := c.createTask(map[string]any{"name": "a", "status": 0, "tags": []string{"x", "y"}})

	var x tagCount
	for _, tg := range c.tags() {
		if tg.Name == "x" {
			x = tg
		}
	}
	require.NotZero(t, x.ID)

	status, _ := c.do(http.MethodDelete, fmt.Sprintf("/tags/%d", x.ID), nil)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, []string{"y"}, c.getTask(id).Tags, "tag detached from the task, task kept")

	status, _ = c.do(http.MethodDelete, fmt.Sprintf("/tags/%d", x.ID), nil)
	assert.Equal(t, http.StatusNotFound, status)
	status, _ = c.do(http.MethodDelete, "/tags/abc", nil)
	assert.Equal(t, http.StatusBadRequest, status)
}

func TestTagsAreIsolatedPerUser(t *testing.T) {
	alice, bob := newUser(t), newUser(t)
	alice.createTask(map[string]any{"name": "a", "status": 0, "tags": []string{"shared-name"}})
	aliceTag := alice.tags()[0]

	assert.Empty(t, bob.tags(), "bob sees no tags")

	status, _ := bob.do(http.MethodDelete, fmt.Sprintf("/tags/%d", aliceTag.ID), nil)
	assert.Equal(t, http.StatusNotFound, status)
	assert.Len(t, alice.tags(), 1, "alice's tag survives")

	// Same name for bob is a separate tag.
	bob.createTask(map[string]any{"name": "b", "status": 0, "tags": []string{"shared-name"}})
	require.Len(t, bob.tags(), 1)
	assert.NotEqual(t, aliceTag.ID, bob.tags()[0].ID)
	assert.EqualValues(t, 1, alice.tags()[0].TaskCount)

	// Bob's tag filter does not match alice's tasks.
	status, env := bob.do(http.MethodGet, "/tasks?tag=shared-name", nil)
	require.Equal(t, http.StatusOK, status)
	assert.EqualValues(t, 1, decode[taskList](t, env).Page.TotalRecords)
}

func TestTagValidation(t *testing.T) {
	c := newUser(t)
	tooMany := make([]string, 11)
	for i := range tooMany {
		tooMany[i] = fmt.Sprintf("t%d", i)
	}
	for name, tags := range map[string][]string{
		"empty tag":   {"ok", "   "},
		"too long":    {strings.Repeat("x", 51)},
		"too many":    tooMany,
		"only spaces": {" "},
	} {
		status, _ := c.do(http.MethodPost, "/tasks", map[string]any{"name": "x", "status": 0, "tags": tags})
		assert.Equal(t, http.StatusBadRequest, status, name)
	}
	status, _ := c.do(http.MethodGet, "/tasks?tag="+strings.Repeat("x", 51), nil)
	assert.Equal(t, http.StatusBadRequest, status)

	// Exactly 10 is fine.
	status, _ = c.do(http.MethodPost, "/tasks", map[string]any{"name": "x", "status": 0, "tags": tooMany[:10]})
	assert.Equal(t, http.StatusCreated, status)
}
