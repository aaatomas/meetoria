package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/meetoria/meetoria/backend/internal/auth/middleware"
	"github.com/meetoria/meetoria/backend/internal/booking"
	bookingservice "github.com/meetoria/meetoria/backend/internal/booking/service"
	apperrors "github.com/meetoria/meetoria/backend/internal/common/errors"
	"github.com/meetoria/meetoria/backend/internal/employee"
	employeerepo "github.com/meetoria/meetoria/backend/internal/employee/repository"
	"github.com/meetoria/meetoria/backend/internal/organization"
	orgservice "github.com/meetoria/meetoria/backend/internal/organization/service"
	servicerepo "github.com/meetoria/meetoria/backend/internal/service/repository"
)

type PublicHandler struct {
	bookingService bookingservice.Service
	orgService     orgservice.Service
	serviceRepo    servicerepo.Repository
	employeeRepo   employeerepo.Repository
}

func NewPublicHandler(
	bookingService bookingservice.Service,
	orgService orgservice.Service,
	serviceRepo servicerepo.Repository,
	employeeRepo employeerepo.Repository,
) *PublicHandler {
	return &PublicHandler{
		bookingService: bookingService,
		orgService:     orgService,
		serviceRepo:    serviceRepo,
		employeeRepo:   employeeRepo,
	}
}

func (h *PublicHandler) GetOrganization(c *gin.Context) {
	org, err := h.resolvePublicOrg(c)
	if err != nil {
		c.Error(err)
		return
	}

	settings := organization.ParseSettings(org.Settings)
	c.JSON(http.StatusOK, booking.PublicOrganizationProfile{
		Name:               org.Name,
		Slug:               org.Slug,
		Timezone:           org.Timezone,
		LogoURL:            org.LogoURL,
		Address:            org.Address,
		Phone:              org.Phone,
		Email:              org.Email,
		CancellationPolicy: settings.Booking.CancellationPolicy,
		ReschedulingPolicy: settings.Booking.ReschedulingPolicy,
		EmailRequired:      settings.Booking.EmailRequired,
		Currency:           organization.NormalizeCurrency(org.Currency),
		TimeFormat:         settings.TimeFormat,
	})
}

func (h *PublicHandler) ListServices(c *gin.Context) {
	org, err := h.resolvePublicOrg(c)
	if err != nil {
		c.Error(err)
		return
	}

	services, _, err := h.serviceRepo.List(c.Request.Context(), org.ID, 0, 1000, true)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, services)
}

func (h *PublicHandler) ListEmployees(c *gin.Context) {
	org, err := h.resolvePublicOrg(c)
	if err != nil {
		c.Error(err)
		return
	}

	serviceIDStr := c.Query("service_id")
	if serviceIDStr == "" {
		c.Error(apperrors.Validation("service_id is required"))
		return
	}
	serviceID, err := uuid.Parse(serviceIDStr)
	if err != nil {
		c.Error(apperrors.Validation("invalid service_id"))
		return
	}

	employees, err := h.employeeRepo.ListByService(c.Request.Context(), org.ID, serviceID)
	if err != nil {
		c.Error(err)
		return
	}
	if len(employees) == 0 {
		all, _, err := h.employeeRepo.List(c.Request.Context(), org.ID, 0, 1000, true)
		if err != nil {
			c.Error(err)
			return
		}
		employees = all
	}

	publicEmployees := make([]employee.PublicEmployee, len(employees))
	for i, e := range employees {
		publicEmployees[i] = employee.PublicEmployee{
			ID:        e.ID,
			FirstName: e.FirstName,
			LastName:  e.LastName,
			Title:     e.Title,
			AvatarURL: e.AvatarURL,
		}
	}

	c.JSON(http.StatusOK, publicEmployees)
}

func (h *PublicHandler) GetAvailability(c *gin.Context) {
	org, err := h.resolvePublicOrg(c)
	if err != nil {
		c.Error(err)
		return
	}

	var req booking.PublicAvailabilityRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(err)
		return
	}

	slots, err := h.bookingService.GetPublicAvailability(
		c.Request.Context(),
		org.ID,
		req,
		organization.ParseSettings(org.Settings).Booking.MinNoticeMinutes,
	)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, slots)
}

func (h *PublicHandler) CreateBooking(c *gin.Context) {
	org, err := h.resolvePublicOrg(c)
	if err != nil {
		c.Error(err)
		return
	}

	var req booking.PublicCreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	correlationID := uuid.New()
	if id, exists := c.Get(middleware.ContextKeyCorrelation); exists {
		if parsed, err := uuid.Parse(id.(string)); err == nil {
			correlationID = parsed
		}
	}

	result, err := h.bookingService.CreatePublic(c.Request.Context(), org, req, correlationID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *PublicHandler) resolvePublicOrg(c *gin.Context) (*organization.Organization, error) {
	slug := c.Param("slug")
	org, err := h.orgService.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		return nil, err
	}

	settings := organization.ParseSettings(org.Settings)
	if !settings.Booking.Enabled {
		return nil, apperrors.NotFound("public booking is not available")
	}

	return org, nil
}
