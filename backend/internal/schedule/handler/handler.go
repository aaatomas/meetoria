package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/meetoria/meetoria/backend/internal/auth/middleware"
	"github.com/meetoria/meetoria/backend/internal/organization"
	orgservice "github.com/meetoria/meetoria/backend/internal/organization/service"
	"github.com/meetoria/meetoria/backend/internal/schedule"
	scheduleservice "github.com/meetoria/meetoria/backend/internal/schedule/service"
	userservice "github.com/meetoria/meetoria/backend/internal/user/service"
)

type Handler struct {
	scheduleService scheduleservice.Service
	orgService      orgservice.Service
	userService     userservice.Service
}

func NewHandler(scheduleService scheduleservice.Service, orgService orgservice.Service, userService userservice.Service) *Handler {
	return &Handler{
		scheduleService: scheduleService,
		orgService:      orgService,
		userService:     userService,
	}
}

func (h *Handler) GetWorkingHours(c *gin.Context) {
	orgID, user, err := h.tenantContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.orgService.VerifyMembership(c.Request.Context(), orgID, user.ID); err != nil {
		c.Error(err)
		return
	}

	schedule, err := h.scheduleService.GetOrgSchedule(c.Request.Context(), orgID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"schedule": schedule})
}

func (h *Handler) SetWorkingHours(c *gin.Context) {
	orgID, user, err := h.tenantContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.orgService.VerifyMembership(c.Request.Context(), orgID, user.ID,
		organization.RoleOrganizationOwner, organization.RoleManager); err != nil {
		c.Error(err)
		return
	}

	var req schedule.SetWorkingHoursRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	if err := h.scheduleService.SetOrgSchedule(c.Request.Context(), orgID, req); err != nil {
		c.Error(err)
		return
	}

	schedule, err := h.scheduleService.GetOrgSchedule(c.Request.Context(), orgID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"schedule": schedule})
}

func (h *Handler) tenantContext(c *gin.Context) (uuid.UUID, *userservice.UserContext, error) {
	orgID, err := uuid.Parse(c.Param("organization_id"))
	if err != nil {
		return uuid.Nil, nil, err
	}

	keycloakID, _ := c.Get(middleware.ContextKeyKeycloakID)
	email, _ := c.Get(middleware.ContextKeyEmail)
	user, err := h.userService.GetOrCreateByKeycloak(c.Request.Context(), keycloakID.(uuid.UUID), email.(string))
	if err != nil {
		return uuid.Nil, nil, err
	}

	return orgID, user, nil
}
