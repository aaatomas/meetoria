package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/meetoria/meetoria/backend/internal/auth/middleware"
	branchservice "github.com/meetoria/meetoria/backend/internal/branch/service"
	apperrors "github.com/meetoria/meetoria/backend/internal/common/errors"
	"github.com/meetoria/meetoria/backend/internal/organization"
	orgservice "github.com/meetoria/meetoria/backend/internal/organization/service"
	"github.com/meetoria/meetoria/backend/internal/schedule"
	scheduleservice "github.com/meetoria/meetoria/backend/internal/schedule/service"
	userservice "github.com/meetoria/meetoria/backend/internal/user/service"
)

type Handler struct {
	scheduleService scheduleservice.Service
	branchService   branchservice.Service
	orgService      orgservice.Service
	userService     userservice.Service
}

func NewHandler(scheduleService scheduleservice.Service, branchService branchservice.Service, orgService orgservice.Service, userService userservice.Service) *Handler {
	return &Handler{scheduleService: scheduleService, branchService: branchService, orgService: orgService, userService: userService}
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
	branchID, err := h.resolveBranchID(c, orgID)
	if err != nil {
		c.Error(err)
		return
	}
	schedule, err := h.scheduleService.GetBranchSchedule(c.Request.Context(), orgID, branchID)
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
	branchID, err := h.resolveBranchIDFromRequest(c, orgID, req.BranchID)
	if err != nil {
		c.Error(err)
		return
	}
	req.BranchID = &branchID
	if err := h.scheduleService.SetBranchSchedule(c.Request.Context(), orgID, req); err != nil {
		c.Error(err)
		return
	}
	schedule, err := h.scheduleService.GetBranchSchedule(c.Request.Context(), orgID, branchID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"schedule": schedule})
}

func (h *Handler) resolveBranchID(c *gin.Context, orgID uuid.UUID) (uuid.UUID, error) {
	if s := c.Query("branch_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			return uuid.Nil, apperrors.Validation("invalid branch_id")
		}
		return h.branchService.ResolveBranchID(c.Request.Context(), orgID, &id)
	}
	if s := c.GetHeader("X-Branch-ID"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			return uuid.Nil, apperrors.Validation("invalid X-Branch-ID header")
		}
		return h.branchService.ResolveBranchID(c.Request.Context(), orgID, &id)
	}
	return h.branchService.ResolveBranchID(c.Request.Context(), orgID, nil)
}

func (h *Handler) resolveBranchIDFromRequest(c *gin.Context, orgID uuid.UUID, branchID *uuid.UUID) (uuid.UUID, error) {
	if branchID != nil {
		return h.branchService.ResolveBranchID(c.Request.Context(), orgID, branchID)
	}
	return h.resolveBranchID(c, orgID)
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
