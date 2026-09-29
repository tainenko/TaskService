package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github/TaskService/middleware"
	"github/TaskService/model"
	"github/TaskService/response"
	"github/TaskService/service"
)

type UserServiceInterface interface {
	Register(ctx context.Context, email, password string) (*model.User, error)
	Authenticate(ctx context.Context, email, password string) (*model.User, error)
}

// TokenIssuer signs access tokens for authenticated users.
type TokenIssuer interface {
	Generate(userID int32) (string, time.Time, error)
}

// FailureLimiter throttles repeated failed logins per account, independent of
// the client IP, so a botnet cannot brute-force one account from many addresses.
type FailureLimiter interface {
	Allowed(key string) (bool, time.Duration)
	Fail(key string)
	Reset(key string)
}

type AuthHandler struct {
	users    UserServiceInterface
	tokens   TokenIssuer
	failures FailureLimiter // optional
}

func NewAuthHandler(users UserServiceInterface, tokens TokenIssuer, failures FailureLimiter) *AuthHandler {
	return &AuthHandler{users: users, tokens: tokens, failures: failures}
}

// bcrypt only hashes the first 72 bytes, so longer passwords are rejected.
type Credentials struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type AuthResponse struct {
	Token     string      `json:"token"`
	ExpiresAt time.Time   `json:"expires_at"`
	User      *model.User `json:"user"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req Credentials
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.InvalidPayload, err.Error())
		return
	}

	user, err := h.users.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			response.Fail(c, http.StatusConflict, response.EmailTaken, err.Error())
			return
		}
		response.Fail(c, http.StatusInternalServerError, response.RegisterErr, "failed to register")
		return
	}

	h.respondWithToken(c, http.StatusCreated, user, response.RegisterErr)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req Credentials
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, response.InvalidPayload, err.Error())
		return
	}

	// Checked before the password so a locked account gives no signal about the password.
	key := strings.ToLower(strings.TrimSpace(req.Email))
	if h.failures != nil {
		if ok, wait := h.failures.Allowed(key); !ok {
			middleware.TooManyRequests(c, wait)
			return
		}
	}

	user, err := h.users.Authenticate(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			if h.failures != nil {
				h.failures.Fail(key)
			}
			response.Fail(c, http.StatusUnauthorized, response.InvalidCredentials, err.Error())
			return
		}
		response.Fail(c, http.StatusInternalServerError, response.LoginErr, "failed to login")
		return
	}

	if h.failures != nil {
		h.failures.Reset(key)
	}
	h.respondWithToken(c, http.StatusOK, user, response.LoginErr)
}

func (h *AuthHandler) respondWithToken(c *gin.Context, status int, user *model.User, errCode response.ErrCode) {
	token, exp, err := h.tokens.Generate(user.ID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, errCode, "failed to issue token")
		return
	}
	c.JSON(status, response.Result{Data: AuthResponse{Token: token, ExpiresAt: exp, User: user}})
}
