package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchResult(t *testing.T) {
	res := newBatchResult([]int32{1, 2, 3, 4}, []int32{4, 2})
	assert.Equal(t, []int32{2, 4}, res.Processed, "keeps request order")
	assert.Equal(t, []int32{1, 3}, res.NotFound)

	empty := newBatchResult([]int32{1}, nil)
	assert.NotNil(t, empty.Processed, "serialized as [] not null")
}

func TestDedupe(t *testing.T) {
	assert.Equal(t, []int32{3, 1, 2}, dedupe([]int32{3, 1, 3, 2, 1}))
}

func idsJSON(n int) string {
	ids := make([]string, n)
	for i := range ids {
		ids[i] = fmt.Sprint(i + 1)
	}
	return "[" + strings.Join(ids, ",") + "]"
}

func TestBatchHandlers_Validation(t *testing.T) {
	h := NewTaskHandler(&MockTaskService{})
	okTask := `{"name":"a","status":0}`

	type tc struct {
		name string
		do   func() int
		want int
	}
	cases := []tc{
		{"create ok", func() int {
			return doRequest(t, h.CreateTasks, http.MethodPost, "/b", "/b", `{"tasks":[`+okTask+`,`+okTask+`]}`).Code
		}, http.StatusCreated},
		{"create empty", func() int {
			return doRequest(t, h.CreateTasks, http.MethodPost, "/b", "/b", `{"tasks":[]}`).Code
		}, http.StatusBadRequest},
		{"create missing", func() int { return doRequest(t, h.CreateTasks, http.MethodPost, "/b", "/b", `{}`).Code }, http.StatusBadRequest},
		{"create too many", func() int {
			items := strings.TrimSuffix(strings.Repeat(okTask+",", MaxBatchSize+1), ",")
			return doRequest(t, h.CreateTasks, http.MethodPost, "/b", "/b", `{"tasks":[`+items+`]}`).Code
		}, http.StatusBadRequest},
		{"create max size ok", func() int {
			items := strings.TrimSuffix(strings.Repeat(okTask+",", MaxBatchSize), ",")
			return doRequest(t, h.CreateTasks, http.MethodPost, "/b", "/b", `{"tasks":[`+items+`]}`).Code
		}, http.StatusCreated},
		{"create invalid item", func() int {
			return doRequest(t, h.CreateTasks, http.MethodPost, "/b", "/b", `{"tasks":[`+okTask+`,{"name":"","status":0}]}`).Code
		}, http.StatusBadRequest},
		{"create invalid tag", func() int {
			return doRequest(t, h.CreateTasks, http.MethodPost, "/b", "/b", `{"tasks":[{"name":"a","status":0,"tags":[" "]}]}`).Code
		}, http.StatusBadRequest},

		{"delete ok", func() int {
			return doRequest(t, h.DeleteTasks, http.MethodDelete, "/b", "/b", `{"ids":[1,2]}`).Code
		}, http.StatusOK},
		{"delete empty", func() int { return doRequest(t, h.DeleteTasks, http.MethodDelete, "/b", "/b", `{"ids":[]}`).Code }, http.StatusBadRequest},
		{"delete non-positive id", func() int {
			return doRequest(t, h.DeleteTasks, http.MethodDelete, "/b", "/b", `{"ids":[1,0]}`).Code
		}, http.StatusBadRequest},
		{"delete too many", func() int {
			return doRequest(t, h.DeleteTasks, http.MethodDelete, "/b", "/b", `{"ids":`+idsJSON(MaxBatchSize+1)+`}`).Code
		}, http.StatusBadRequest},

		{"status ok", func() int {
			return doRequest(t, h.UpdateTasksStatus, http.MethodPatch, "/b", "/b", `{"ids":[1],"status":0}`).Code
		}, http.StatusOK},
		{"status missing", func() int {
			return doRequest(t, h.UpdateTasksStatus, http.MethodPatch, "/b", "/b", `{"ids":[1]}`).Code
		}, http.StatusBadRequest},
		{"status invalid", func() int {
			return doRequest(t, h.UpdateTasksStatus, http.MethodPatch, "/b", "/b", `{"ids":[1],"status":9}`).Code
		}, http.StatusBadRequest},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, c.do())
		})
	}
}

func TestBatchCreate_ReturnsIDsInOrder(t *testing.T) {
	h := NewTaskHandler(&MockTaskService{})
	w := doRequest(t, h.CreateTasks, http.MethodPost, "/b", "/b", `{"tasks":[{"name":"a","status":0},{"name":"b","status":1}]}`)
	require.Equal(t, http.StatusCreated, w.Code)
	var body struct {
		Data struct {
			IDs []int32 `json:"ids"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, []int32{1, 2}, body.Data.IDs)
}
