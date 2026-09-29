package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github/TaskService/response"
)

// Pinger reports whether a dependency (e.g. the database) is reachable.
type Pinger interface {
	PingContext(ctx context.Context) error
}

type HealthHandler struct {
	db Pinger
}

func NewHealthHandler(db Pinger) *HealthHandler {
	return &HealthHandler{db: db}
}

// Healthz is the liveness probe: the process is up and serving requests.
func (h *HealthHandler) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Readyz is the readiness probe: the service can reach its database.
func (h *HealthHandler) Readyz(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		response.Fail(c, http.StatusServiceUnavailable, response.ServerError, "database not ready")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}
