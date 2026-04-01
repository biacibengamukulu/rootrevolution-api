package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	appuser "rootrevolution-api/internal/application/user"
	"rootrevolution-api/internal/domain/user"
)

type AuthMiddleware struct {
	userSvc *appuser.Service
}

func NewAuthMiddleware(userSvc *appuser.Service) *AuthMiddleware {
	return &AuthMiddleware{userSvc: userSvc}
}

func (m *AuthMiddleware) Required() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header required",
			})
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid Authorization header format. Use: Bearer <token>",
			})
		}

		claims, err := m.userSvc.ValidateToken(parts[1])
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		c.Locals("userID", claims.UserID)
		c.Locals("userEmail", claims.Email)
		c.Locals("userRole", claims.Role)
		c.Locals("userName", claims.Name)
		c.Locals("userSurname", claims.Surname)
		c.Locals("claims", claims)

		return c.Next()
	}
}

func (m *AuthMiddleware) AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("userRole").(user.Role)
		if !ok || role != user.RoleAdmin {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Admin access required",
			})
		}
		return c.Next()
	}
}
