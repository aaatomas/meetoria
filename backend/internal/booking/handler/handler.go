package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/meetoria/meetoria/backend/internal/auth/middleware"
	"github.com/meetoria/meetoria/backend/internal/booking"
	bookingservice "github.com/meetoria/meetoria/backend/internal/booking/service"
	bookingrepo "github.com/meetoria/meetoria/backend/internal/booking/repository"
	apperrors "github.com/meetoria/meetoria/backend/internal/common/errors"
	commonmodel "github.com/meetoria/meetoria/backend/internal/common/model"
	"github.com/meetoria/meetoria/backend/internal/organization"
	orgservice "github.com/meetoria/meetoria/backend/internal/organization/service"
	userservice "github.com/meetoria/meetoria/backend/internal/user/service"
)

type Handler struct {
	bookingService bookingservice.Service
	orgService     orgservice.Service
	userService    userservice.Service
}

func NewHandler(bookingService bookingservice.Service, orgService orgservice.Service, userService userservice.Service) *Handler {
	return &Handler{bookingService: bookingService, orgService: orgService, userService: userService}
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

	var req booking.CreateBookingRequest
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

	correlationID := h.correlationID(c)
	result, err := h.bookingService.Create(c.Request.Context(), orgID, req, correlationID)
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

	id, _ := uuid.Parse(c.Param("booking_id"))
	result, err := h.bookingService.GetByID(c.Request.Context(), orgID, id)
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

	id, _ := uuid.Parse(c.Param("booking_id"))
	var req booking.UpdateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := h.bookingService.Update(c.Request.Context(), orgID, id, req, h.correlationID(c))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) Cancel(c *gin.Context) {
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

	id, _ := uuid.Parse(c.Param("booking_id"))
	var req booking.CancelBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := h.bookingService.Cancel(c.Request.Context(), orgID, id, req, h.correlationID(c))
	if err != nil {
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

	filters := bookingrepo.ListFilters{}
	if branchID, err := h.resolveBranchFilter(c); err != nil {
		c.Error(err)
		return
	} else if branchID != nil {
		filters.BranchID = branchID
	}
	if empID := c.Query("employee_id"); empID != "" {
		id, _ := uuid.Parse(empID)
		filters.EmployeeID = &id
	}
	if custID := c.Query("customer_id"); custID != "" {
		id, _ := uuid.Parse(custID)
		filters.CustomerID = &id
	}

	result, err := h.bookingService.List(c.Request.Context(), orgID, filters, params)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetAvailability(c *gin.Context) {
	orgID, user, err := h.tenantContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.orgService.VerifyMembership(c.Request.Context(), orgID, user.ID); err != nil {
		c.Error(err)
		return
	}

	var req booking.AvailabilityRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(err)
		return
	}

	slots, err := h.bookingService.GetAvailability(c.Request.Context(), orgID, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, slots)
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

	id, _ := uuid.Parse(c.Param("booking_id"))
	if err := h.bookingService.SendSMS(c.Request.Context(), orgID, id, h.correlationID(c)); err != nil {
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

	id, _ := uuid.Parse(c.Param("booking_id"))
	if err := h.bookingService.SendEmail(c.Request.Context(), orgID, id, h.correlationID(c)); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) ListNotifications(c *gin.Context) {
	orgID, user, err := h.tenantContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.orgService.VerifyMembership(c.Request.Context(), orgID, user.ID); err != nil {
		c.Error(err)
		return
	}

	id, _ := uuid.Parse(c.Param("booking_id"))
	notifications, err := h.bookingService.ListNotifications(c.Request.Context(), orgID, id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, notifications)
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

func (h *Handler) correlationID(c *gin.Context) uuid.UUID {
	if id, exists := c.Get(middleware.ContextKeyCorrelation); exists {
		if parsed, err := uuid.Parse(id.(string)); err == nil {
			return parsed
		}
	}
	return uuid.New()
}
