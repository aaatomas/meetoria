package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/meetoria/meetoria/backend/internal/auth/middleware"
	"github.com/meetoria/meetoria/backend/internal/user"
	userservice "github.com/meetoria/meetoria/backend/internal/user/service"
)

type Handler struct {
	userService userservice.Service
}

func NewHandler(userService userservice.Service) *Handler {
	return &Handler{userService: userService}
}

func (h *Handler) GetMe(c *gin.Context) {
	keycloakID, _ := c.Get(middleware.ContextKeyKeycloakID)
	email, _ := c.Get(middleware.ContextKeyEmail)

	ctx, err := h.userService.GetOrCreateByKeycloak(c.Request.Context(), keycloakID.(uuid.UUID), email.(string))
	if err != nil {
		c.Error(err)
		return
	}

	u, err := h.userService.GetByID(c.Request.Context(), ctx.ID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, u)
}

func (h *Handler) UpdateMe(c *gin.Context) {
	keycloakID, _ := c.Get(middleware.ContextKeyKeycloakID)
	email, _ := c.Get(middleware.ContextKeyEmail)

	ctx, err := h.userService.GetOrCreateByKeycloak(c.Request.Context(), keycloakID.(uuid.UUID), email.(string))
	if err != nil {
		c.Error(err)
		return
	}

	var req user.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	u, err := h.userService.Update(c.Request.Context(), ctx.ID, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, u)
}

func (h *Handler) SyncFromKeycloak(c *gin.Context) {
	keycloakID, _ := c.Get(middleware.ContextKeyKeycloakID)

	var req user.SyncUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	u, err := h.userService.SyncFromKeycloak(c.Request.Context(), keycloakID.(uuid.UUID), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, u)
}
