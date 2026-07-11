package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"github.com/meetoria/meetoria/backend/internal/analytics/handler"
	analyticsservice "github.com/meetoria/meetoria/backend/internal/analytics/service"
	analyticrepo "github.com/meetoria/meetoria/backend/internal/analytics/repository"
	"github.com/meetoria/meetoria/backend/internal/auth/keycloak"
	"github.com/meetoria/meetoria/backend/internal/auth/middleware"
	bookinghandler "github.com/meetoria/meetoria/backend/internal/booking/handler"
	bookingservice "github.com/meetoria/meetoria/backend/internal/booking/service"
	bookingrepo "github.com/meetoria/meetoria/backend/internal/booking/repository"
	"github.com/meetoria/meetoria/backend/internal/common/config"
	redisclient "github.com/meetoria/meetoria/backend/internal/common/redis"
	"github.com/meetoria/meetoria/backend/internal/common/rabbitmq"
	"github.com/meetoria/meetoria/backend/internal/common/storage"
	customerhandler "github.com/meetoria/meetoria/backend/internal/customer/handler"
	customerservice "github.com/meetoria/meetoria/backend/internal/customer/service"
	customerrepo "github.com/meetoria/meetoria/backend/internal/customer/repository"
	employeehandler "github.com/meetoria/meetoria/backend/internal/employee/handler"
	employeeservice "github.com/meetoria/meetoria/backend/internal/employee/service"
	employeerepo "github.com/meetoria/meetoria/backend/internal/employee/repository"
	orghandler "github.com/meetoria/meetoria/backend/internal/organization/handler"
	orgservice "github.com/meetoria/meetoria/backend/internal/organization/service"
	orgrepo "github.com/meetoria/meetoria/backend/internal/organization/repository"
	servicehandler "github.com/meetoria/meetoria/backend/internal/service/handler"
	serviceservice "github.com/meetoria/meetoria/backend/internal/service/service"
	servicerepo "github.com/meetoria/meetoria/backend/internal/service/repository"
	notifservice "github.com/meetoria/meetoria/backend/internal/notification/service"
	notifrepo "github.com/meetoria/meetoria/backend/internal/notification/repository"
	schedulerepo "github.com/meetoria/meetoria/backend/internal/schedule/repository"
	userhandler "github.com/meetoria/meetoria/backend/internal/user/handler"
	userservice "github.com/meetoria/meetoria/backend/internal/user/service"
	userrepo "github.com/meetoria/meetoria/backend/internal/user/repository"
)

type Dependencies struct {
	Config    *config.Config
	DB        *gorm.DB
	Redis     *redisclient.Client
	Publisher *rabbitmq.Publisher
}

func Setup(deps Dependencies) *gin.Engine {
	if !deps.Config.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.CorrelationID())
	r.Use(middleware.StructuredLogger())
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.RateLimit(deps.Config.RateLimit.RequestsPerMinute))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "meetoria-api"})
	})

	fileStorage, err := storage.NewLocalStorage(deps.Config.UploadDir)
	if err != nil {
		panic(err)
	}
	r.Static("/uploads", deps.Config.UploadDir)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	validator := keycloak.NewTokenValidator(
		deps.Config.Keycloak.JWTIssuer,
		deps.Config.Keycloak.URL,
		deps.Config.Keycloak.Realm,
	)

	userRepo := userrepo.NewRepository(deps.DB)
	userSvc := userservice.NewService(userRepo)

	orgRepo := orgrepo.NewRepository(deps.DB)
	orgSvc := orgservice.NewService(orgRepo)

	customerRepo := customerrepo.NewRepository(deps.DB)
	employeeRepo := employeerepo.NewRepository(deps.DB)
	employeeSvc := employeeservice.NewService(employeeRepo, fileStorage)

	serviceRepo := servicerepo.NewRepository(deps.DB)
	serviceSvc := serviceservice.NewService(serviceRepo)

	scheduleRepo := schedulerepo.NewRepository(deps.DB)
	notifRepo := notifrepo.NewRepository(deps.DB)
	notifSvc := notifservice.NewService(notifRepo, customerRepo, employeeRepo, deps.Publisher)

	bookingRepo := bookingrepo.NewRepository(deps.DB)
	customerSvc := customerservice.NewService(customerRepo, bookingRepo, notifSvc)
	bookingSvc := bookingservice.NewService(
		bookingRepo, customerRepo, employeeRepo, serviceRepo,
		scheduleRepo, deps.Redis, deps.Publisher, notifSvc,
	)

	analyticsRepo := analyticrepo.NewRepository(deps.DB)
	analyticsSvc := analyticsservice.NewService(analyticsRepo, deps.Redis)

	orgHandler := orghandler.NewHandler(orgSvc, userSvc)
	userHandler := userhandler.NewHandler(userSvc)
	customerHandler := customerhandler.NewHandler(customerSvc, orgSvc, userSvc)
	employeeHandler := employeehandler.NewHandler(employeeSvc, orgSvc, userSvc, fileStorage)
	serviceHandler := servicehandler.NewHandler(serviceSvc, orgSvc, userSvc)
	bookingHandler := bookinghandler.NewHandler(bookingSvc, orgSvc, userSvc)
	analyticsHandler := handler.NewHandler(analyticsSvc, orgSvc, userSvc)

	api := r.Group("/api/v1")
	api.Use(middleware.JWTAuth(validator))

	api.GET("/me", userHandler.GetMe)
	api.PUT("/me", userHandler.UpdateMe)
	api.POST("/me/sync", userHandler.SyncFromKeycloak)

	api.POST("/organizations", orgHandler.CreateOrganization)
	api.GET("/organizations", orgHandler.ListOrganizations)

	orgRoutes := api.Group("/organizations/:organization_id")
	orgRoutes.Use(middleware.OrganizationContext())
	{
		orgRoutes.GET("", orgHandler.GetOrganization)
		orgRoutes.PUT("", orgHandler.UpdateOrganization)

		orgRoutes.GET("/customers", customerHandler.List)
		orgRoutes.POST("/customers", customerHandler.Create)
		orgRoutes.GET("/customers/:customer_id", customerHandler.Get)
		orgRoutes.PUT("/customers/:customer_id", customerHandler.Update)
		orgRoutes.DELETE("/customers/:customer_id", customerHandler.Delete)
		orgRoutes.POST("/customers/:customer_id/notifications/sms", customerHandler.SendSMS)
		orgRoutes.POST("/customers/:customer_id/notifications/email", customerHandler.SendEmail)

		orgRoutes.GET("/employees", employeeHandler.List)
		orgRoutes.POST("/employees", employeeHandler.Create)
		orgRoutes.GET("/employees/:employee_id", employeeHandler.Get)
		orgRoutes.PUT("/employees/:employee_id", employeeHandler.Update)
		orgRoutes.POST("/employees/:employee_id/avatar", employeeHandler.UploadAvatar)
		orgRoutes.DELETE("/employees/:employee_id", employeeHandler.Delete)

		orgRoutes.GET("/services", serviceHandler.List)
		orgRoutes.POST("/services", serviceHandler.Create)
		orgRoutes.GET("/services/:service_id", serviceHandler.Get)
		orgRoutes.PUT("/services/:service_id", serviceHandler.Update)
		orgRoutes.DELETE("/services/:service_id", serviceHandler.Delete)

		orgRoutes.GET("/bookings", bookingHandler.List)
		orgRoutes.POST("/bookings", bookingHandler.Create)
		orgRoutes.GET("/bookings/availability", bookingHandler.GetAvailability)
		orgRoutes.GET("/bookings/:booking_id", bookingHandler.Get)
		orgRoutes.PUT("/bookings/:booking_id", bookingHandler.Update)
		orgRoutes.POST("/bookings/:booking_id/cancel", bookingHandler.Cancel)
		orgRoutes.POST("/bookings/:booking_id/notifications/sms", bookingHandler.SendSMS)
		orgRoutes.POST("/bookings/:booking_id/notifications/email", bookingHandler.SendEmail)

		orgRoutes.GET("/analytics/dashboard", analyticsHandler.GetDashboard)
		orgRoutes.GET("/analytics/employees/:employee_id", analyticsHandler.GetEmployeeAnalytics)
		orgRoutes.GET("/analytics/customers/:customer_id", analyticsHandler.GetCustomerAnalytics)
	}

	return r
}
