package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/meetoria/meetoria/backend/internal/auth/middleware"
	"github.com/meetoria/meetoria/backend/internal/customer"
	customerservice "github.com/meetoria/meetoria/backend/internal/customer/service"
	commonmodel "github.com/meetoria/meetoria/backend/internal/common/model"
	"github.com/meetoria/meetoria/backend/internal/organization"
	orgservice "github.com/meetoria/meetoria/backend/internal/organization/service"
	userservice "github.com/meetoria/meetoria/backend/internal/user/service"
)

type Handler struct {
	customerService customerservice.Service
	orgService      orgservice.Service
	userService     userservice.Service
}

func NewHandler(customerService customerservice.Service, orgService orgservice.Service, userService userservice.Service) *Handler {
	return &Handler{customerService: customerService, orgService: orgService, userService: userService}
}

func (h *Handler) Create(c *gin.Context) {
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

	var req customer.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := h.customerService.Create(c.Request.Context(), orgID, req)
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

	id, _ := uuid.Parse(c.Param("customer_id"))
	result, err := h.customerService.GetByID(c.Request.Context(), orgID, id)
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
		organization.RoleOrganizationOwner, organization.RoleManager, organization.RoleEmployee); err != nil {
		c.Error(err)
		return
	}

	id, _ := uuid.Parse(c.Param("customer_id"))
	var req customer.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := h.customerService.Update(c.Request.Context(), orgID, id, req)
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

	id, _ := uuid.Parse(c.Param("customer_id"))
	check, err := h.customerService.CheckDeletion(c.Request.Context(), orgID, id)
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

	id, _ := uuid.Parse(c.Param("customer_id"))
	if err := h.customerService.Delete(c.Request.Context(), orgID, id); err != nil {
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
	search := c.Query("search")

	result, err := h.customerService.List(c.Request.Context(), orgID, params, search)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) SendSMS(c *gin.Context) {
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

	id, _ := uuid.Parse(c.Param("customer_id"))
	if err := h.customerService.SendSMS(c.Request.Context(), orgID, id, h.correlationID(c)); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) SendEmail(c *gin.Context) {
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

	id, _ := uuid.Parse(c.Param("customer_id"))
	if err := h.customerService.SendEmail(c.Request.Context(), orgID, id, h.correlationID(c)); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
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

func (h *Handler) correlationID(c *gin.Context) uuid.UUID {
	if id, exists := c.Get(middleware.ContextKeyCorrelation); exists {
		if parsed, err := uuid.Parse(id.(string)); err == nil {
			return parsed
		}
	}
	return uuid.New()
}
