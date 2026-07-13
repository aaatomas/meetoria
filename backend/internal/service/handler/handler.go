package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/meetoria/meetoria/backend/internal/auth/middleware"
	apperrors "github.com/meetoria/meetoria/backend/internal/common/errors"
	commonmodel "github.com/meetoria/meetoria/backend/internal/common/model"
	"github.com/meetoria/meetoria/backend/internal/organization"
	orgservice "github.com/meetoria/meetoria/backend/internal/organization/service"
	svc "github.com/meetoria/meetoria/backend/internal/service"
	serviceservice "github.com/meetoria/meetoria/backend/internal/service/service"
	userservice "github.com/meetoria/meetoria/backend/internal/user/service"
)

type Handler struct {
	serviceService serviceservice.Service
	orgService     orgservice.Service
	userService    userservice.Service
}

func NewHandler(serviceService serviceservice.Service, orgService orgservice.Service, userService userservice.Service) *Handler {
	return &Handler{serviceService: serviceService, orgService: orgService, userService: userService}
}

func (h *Handler) Create(c *gin.Context) {
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

	var req svc.CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	if req.BranchID == nil || *req.BranchID == uuid.Nil {
		if hdr := c.GetHeader("X-Branch-ID"); hdr != "" {
			if id, err := uuid.Parse(hdr); err == nil {
				req.BranchID = &id
			}
		}
	}

	org, err := h.orgService.GetByID(c.Request.Context(), orgID)
	if err != nil {
		c.Error(err)
		return
	}

	result, err := h.serviceService.Create(c.Request.Context(), orgID, req, org.Currency)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *Handler) Get(c *gin.Context) {
	orgID, user, err := h.tenantContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.orgService.VerifyMembership(c.Request.Context(), orgID, user.ID); err != nil {
		c.Error(err)
		return
	}

	id, _ := uuid.Parse(c.Param("service_id"))
	result, err := h.serviceService.GetByID(c.Request.Context(), orgID, id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) Update(c *gin.Context) {
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

	id, _ := uuid.Parse(c.Param("service_id"))
	var req svc.UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := h.serviceService.Update(c.Request.Context(), orgID, id, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) CheckDeletion(c *gin.Context) {
	orgID, user, err := h.tenantContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.orgService.VerifyMembership(c.Request.Context(), orgID, user.ID); err != nil {
		c.Error(err)
		return
	}

	id, _ := uuid.Parse(c.Param("service_id"))
	check, err := h.serviceService.CheckDeletion(c.Request.Context(), orgID, id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, check)
}

func (h *Handler) Delete(c *gin.Context) {
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

	id, _ := uuid.Parse(c.Param("service_id"))
	if err := h.serviceService.Delete(c.Request.Context(), orgID, id); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) List(c *gin.Context) {
	orgID, user, err := h.tenantContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.orgService.VerifyMembership(c.Request.Context(), orgID, user.ID); err != nil {
		c.Error(err)
		return
	}

	var params commonmodel.PaginationParams
	_ = c.ShouldBindQuery(&params)
	activeOnly := c.Query("active_only") == "true"

	branchID, err := h.resolveBranchFilter(c)
	if err != nil {
		c.Error(err)
		return
	}

	result, err := h.serviceService.List(c.Request.Context(), orgID, branchID, params, activeOnly)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) resolveBranchFilter(c *gin.Context) (*uuid.UUID, error) {
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
