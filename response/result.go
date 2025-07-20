// Package response provides HTTP middleware components for the task service.
package response

import "github.com/gin-gonic/gin"

type Result struct {
	ctx *gin.Context
}

func NewResult(ctx *gin.Context) *Result {
	return &Result{ctx: ctx}
}
