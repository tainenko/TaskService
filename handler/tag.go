package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github/TaskService/middleware"
	"github/TaskService/response"
	"github/TaskService/service"
)

type TagServiceInterface interface {
	ListTags(ctx context.Context, userID int32) ([]service.TagCount, error)
	DeleteTag(ctx context.Context, userID, id int32) error
}

type TagHandler struct {
	tags TagServiceInterface
}

func NewTagHandler(tags TagServiceInterface) *TagHandler {
	return &TagHandler{tags: tags}
}

func (h *TagHandler) ListTags(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.Unauthorized, "unauthorized")
		return
	}

	tags, err := h.tags.ListTags(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, response.GetTagsErr, "failed to list tags")
		return
	}
	response.Success(c, tags)
}

func (h *TagHandler) DeleteTag(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, response.Unauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil || id <= 0 {
		response.Fail(c, http.StatusBadRequest, response.InvalidParam, "invalid tag id")
		return
	}

	if err := h.tags.DeleteTag(c.Request.Context(), userID, int32(id)); err != nil {
		if errors.Is(err, service.ErrTagNotFound) {
			response.Fail(c, http.StatusNotFound, response.TagNotFound, err.Error())
			return
		}
		response.Fail(c, http.StatusInternalServerError, response.DeleteTagErr, "failed to delete tag")
		return
	}
	response.Success(c, "Tag deleted successfully")
}
