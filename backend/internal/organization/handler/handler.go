package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/meetoria/meetoria/backend/internal/auth/middleware"
	commonmodel "github.com/meetoria/meetoria/backend/internal/common/model"
	"github.com/meetoria/meetoria/backend/internal/organization"
	orgservice "github.com/meetoria/meetoria/backend/internal/organization/service"
	userservice "github.com/meetoria/meetoria/backend/internal/user/service"
)

type Handler struct {
	orgService  orgservice.Service
	userService userservice.Service
}

func NewHandler(orgService orgservice.Service, userService userservice.Service) *Handler {
	return &Handler{orgService: orgService, userService: userService}
}

// CreateOrganization godoc
// @Summary Create a new organization
// @Tags organizations
// @Accept json
// @Produce json
// @Param body body organization.CreateOrganizationRequest true "Organization data"
// @Success 201 {object} organization.Organization
// @Router /api/v1/organizations [post]
func (h *Handler) CreateOrganization(c *gin.Context) {
	var req organization.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	user, err := h.resolveUser(c)
	if err != nil {
		c.Error(err)
		return
	}

	org, err := h.orgService.Create(c.Request.Context(), req, user.ID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, org)
}

// ListOrganizations godoc
// @Summary List user's organizations
// @Tags organizations
// @Produce json
// @Success 200 {object} commonmodel.PaginatedResponse[organization.Organization]
// @Router /api/v1/organizations [get]
func (h *Handler) ListOrganizations(c *gin.Context) {
	var params commonmodel.PaginationParams
	_ = c.ShouldBindQuery(&params)
	params.Normalize()

	user, err := h.resolveUser(c)
	if err != nil {
		c.Error(err)
		return
	}

	result, err := h.orgService.List(c.Request.Context(), user.ID, params)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetOrganization godoc
// @Summary Get organization by ID
// @Tags organizations
// @Produce json
// @Param organization_id path string true "Organization ID"
// @Success 200 {object} organization.Organization
// @Router /api/v1/organizations/{organization_id} [get]
func (h *Handler) GetOrganization(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("organization_id"))
	if err != nil {
		c.Error(err)
		return
	}

	user, err := h.resolveUser(c)
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.orgService.VerifyMembership(c.Request.Context(), orgID, user.ID); err != nil {
		c.Error(err)
		return
	}

	org, err := h.orgService.GetByID(c.Request.Context(), orgID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, org)
}

func (h *Handler) UpdateOrganization(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("organization_id"))
	if err != nil {
		c.Error(err)
		return
	}

	user, err := h.resolveUser(c)
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.orgService.VerifyMembership(c.Request.Context(), orgID, user.ID,
		organization.RoleOrganizationOwner, organization.RoleManager); err != nil {
		c.Error(err)
		return
	}

	var req organization.UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	org, err := h.orgService.Update(c.Request.Context(), orgID, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, org)
}

func (h *Handler) resolveUser(c *gin.Context) (*userservice.UserContext, error) {
	keycloakID, _ := c.Get(middleware.ContextKeyKeycloakID)
	email, _ := c.Get(middleware.ContextKeyEmail)

	return h.userService.GetOrCreateByKeycloak(c.Request.Context(), keycloakID.(uuid.UUID), email.(string))
}
