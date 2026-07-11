package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/meetoria/meetoria/backend/internal/auth/keycloak"
	apperrors "github.com/meetoria/meetoria/backend/internal/common/errors"
	"github.com/meetoria/meetoria/backend/internal/common/logger"
)

const (
	ContextKeyUserID      = "user_id"
	ContextKeyKeycloakID  = "keycloak_id"
	ContextKeyEmail       = "email"
	ContextKeyRoles       = "roles"
	ContextKeyOrgID       = "organization_id"
	ContextKeyRequestID   = "request_id"
	ContextKeyCorrelation = "correlation_id"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set(ContextKeyRequestID, requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func CorrelationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}
		c.Set(ContextKeyCorrelation, correlationID)
		c.Header("X-Correlation-ID", correlationID)
		c.Next()
	}
}

func JWTAuth(validator *keycloak.TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "missing authorization header",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "invalid authorization header format",
			})
			return
		}

		claims, err := validator.ValidateToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "invalid or expired token",
			})
			return
		}

		keycloakID, err := uuid.Parse(claims.Subject)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "invalid token subject",
			})
			return
		}

		c.Set(ContextKeyKeycloakID, keycloakID)
		c.Set(ContextKeyEmail, claims.Email)
		c.Set(ContextKeyRoles, claims.RealmAccess.Roles)
		c.Next()
	}
}

func OrganizationContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		orgIDStr := c.GetHeader("X-Organization-ID")
		if orgIDStr == "" {
			orgIDStr = c.Param("organization_id")
		}
		if orgIDStr != "" {
			orgID, err := uuid.Parse(orgIDStr)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"code":    "VALIDATION_ERROR",
					"message": "invalid organization id",
				})
				return
			}
			c.Set(ContextKeyOrgID, orgID)
		}
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get(ContextKeyRoles)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    "FORBIDDEN",
				"message": "insufficient permissions",
			})
			return
		}

		roleList, ok := userRoles.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    "FORBIDDEN",
				"message": "insufficient permissions",
			})
			return
		}

		for _, required := range roles {
			for _, userRole := range roleList {
				if userRole == required {
					c.Next()
					return
				}
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"code":    "FORBIDDEN",
			"message": "insufficient permissions",
		})
	}
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}

		err := c.Errors.Last().Err
		appErr := apperrors.MapError(err)
		c.JSON(appErr.Status, gin.H{
			"code":    appErr.Code,
			"message": appErr.Message,
		})
	}
}

func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		log := logger.Default().With(
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
		)

		if requestID, exists := c.Get(ContextKeyRequestID); exists {
			log = log.With("request_id", requestID)
		}
		if correlationID, exists := c.Get(ContextKeyCorrelation); exists {
			log = log.With("correlation_id", correlationID)
		}
		if orgID, exists := c.Get(ContextKeyOrgID); exists {
			log = log.With("organization_id", orgID)
		}

		log.Info("request completed")
	}
}
