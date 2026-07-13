package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/meetoria/meetoria/backend/internal/auth/middleware"
	"github.com/meetoria/meetoria/backend/internal/employee"
	employeeservice "github.com/meetoria/meetoria/backend/internal/employee/service"
	apperrors "github.com/meetoria/meetoria/backend/internal/common/errors"
	commonmodel "github.com/meetoria/meetoria/backend/internal/common/model"
	"github.com/meetoria/meetoria/backend/internal/common/storage"
	"github.com/meetoria/meetoria/backend/internal/organization"
	orgservice "github.com/meetoria/meetoria/backend/internal/organization/service"
	userservice "github.com/meetoria/meetoria/backend/internal/user/service"
)

type Handler struct {
	employeeService employeeservice.Service
	orgService      orgservice.Service
	userService     userservice.Service
	storage         *storage.LocalStorage
}

func NewHandler(employeeService employeeservice.Service, orgService orgservice.Service, userService userservice.Service, fileStorage *storage.LocalStorage) *Handler {
	return &Handler{employeeService: employeeService, orgService: orgService, userService: userService, storage: fileStorage}
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

	var req employee.CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := h.employeeService.Create(c.Request.Context(), orgID, req)
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

	id, _ := uuid.Parse(c.Param("employee_id"))
	result, err := h.employeeService.GetByID(c.Request.Context(), orgID, id)
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

	id, _ := uuid.Parse(c.Param("employee_id"))
	var req employee.UpdateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := h.employeeService.Update(c.Request.Context(), orgID, id, req)
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

	id, _ := uuid.Parse(c.Param("employee_id"))
	check, err := h.employeeService.CheckDeletion(c.Request.Context(), orgID, id)
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

	id, _ := uuid.Parse(c.Param("employee_id"))
	if err := h.employeeService.Delete(c.Request.Context(), orgID, id); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) UploadAvatar(c *gin.Context) {
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

	id, _ := uuid.Parse(c.Param("employee_id"))
	file, err := c.FormFile("avatar")
	if err != nil {
		c.Error(apperrors.Validation("avatar file is required"))
		return
	}

	avatarURL, err := h.storage.SaveEmployeeAvatar(orgID.String(), id.String(), file)
	if err != nil {
		c.Error(err)
		return
	}

	result, err := h.employeeService.UpdateAvatar(c.Request.Context(), orgID, id, avatarURL)
	if err != nil {
		_ = h.storage.DeleteByURL(avatarURL)
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
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

	result, err := h.employeeService.List(c.Request.Context(), orgID, params, activeOnly)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
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
