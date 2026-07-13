package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	analyticsservice "github.com/meetoria/meetoria/backend/internal/analytics/service"
	"github.com/meetoria/meetoria/backend/internal/auth/middleware"
	branchservice "github.com/meetoria/meetoria/backend/internal/branch/service"
	apperrors "github.com/meetoria/meetoria/backend/internal/common/errors"
	"github.com/meetoria/meetoria/backend/internal/organization"
	orgservice "github.com/meetoria/meetoria/backend/internal/organization/service"
	userservice "github.com/meetoria/meetoria/backend/internal/user/service"
)

type Handler struct {
	analyticsService analyticsservice.Service
	orgService       orgservice.Service
	branchService    branchservice.Service
	userService      userservice.Service
}

func NewHandler(
	analyticsService analyticsservice.Service,
	orgService orgservice.Service,
	branchService branchservice.Service,
	userService userservice.Service,
) *Handler {
	return &Handler{
		analyticsService: analyticsService,
		orgService:       orgService,
		branchService:    branchService,
		userService:      userService,
	}
}

func (h *Handler) GetDashboard(c *gin.Context) {
	orgID, user, err := h.tenantContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.orgService.VerifyMembership(c.Request.Context(), orgID, user.ID,
		organization.RoleOrganizationOwner, organization.RoleManager, organization.RoleEmployee); err != nil {
		c.Error(err)
		return
	}

	branchID, err := h.resolveBranchFilter(c)
	if err != nil {
		c.Error(err)
		return
	}

	from, to := h.parseDateRange(c)

	dashboard, err := h.analyticsService.GetDashboard(c.Request.Context(), orgID, branchID, from, to)
	if err != nil {
		c.Error(err)
		return
	}

	if branchID != nil {
		branch, branchErr := h.branchService.GetByID(c.Request.Context(), orgID, *branchID)
		if branchErr == nil {
			dashboard.BranchName = branch.Name
		}
	}

	c.JSON(http.StatusOK, dashboard)
}

func (h *Handler) GetEmployeeAnalytics(c *gin.Context) {
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

	employeeID, _ := uuid.Parse(c.Param("employee_id"))
	from, to := h.parseDateRange(c)

	stats, err := h.analyticsService.GetEmployeeAnalytics(c.Request.Context(), orgID, employeeID, from, to)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *Handler) GetCustomerAnalytics(c *gin.Context) {
	orgID, user, err := h.tenantContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.orgService.VerifyMembership(c.Request.Context(), orgID, user.ID); err != nil {
		c.Error(err)
		return
	}

	customerID, _ := uuid.Parse(c.Param("customer_id"))
	stats, err := h.analyticsService.GetCustomerAnalytics(c.Request.Context(), orgID, customerID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *Handler) resolveBranchFilter(c *gin.Context) (*uuid.UUID, error) {
	if c.Query("scope") == "organization" {
		return nil, nil
	}

	if s := c.Query("branch_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, apperrors.Validation("invalid branch_id")
		}
		return &id, nil
	}
	if s := c.GetHeader("X-Branch-ID"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, apperrors.Validation("invalid X-Branch-ID header")
		}
		return &id, nil
	}
	return nil, nil
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

func (h *Handler) parseDateRange(c *gin.Context) (time.Time, time.Time) {
	now := time.Now().UTC()
	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 999999999, time.UTC)

	if fromStr := c.Query("from"); fromStr != "" {
		if parsed, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if parsed, err := time.Parse("2006-01-02", toStr); err == nil {
			to = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, time.UTC)
		}
	}

	return from, to
}
