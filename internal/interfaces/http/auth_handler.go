package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appuser "rootrevolution-api/internal/application/user"
	"rootrevolution-api/internal/domain/user"
)

type AuthHandler struct {
	userSvc *appuser.Service
}

func NewAuthHandler(userSvc *appuser.Service) *AuthHandler {
	return &AuthHandler{userSvc: userSvc}
}

// Login godoc
// POST /backend_rootrevolution/api/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req appuser.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email and password are required",
		})
	}

	token, u, err := h.userSvc.Login(req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":      u.ID,
			"name":    u.Name,
			"surname": u.Surname,
			"email":   u.Email,
			"role":    u.Role,
			"status":  u.Status,
		},
	})
}

// Register godoc
// POST /backend_rootrevolution/api/auth/register
// Requires admin
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req appuser.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name, email and password are required",
		})
	}

	requesterRole, _ := c.Locals("userRole").(user.Role)

	u, err := h.userSvc.Register(req, requesterRole)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
		"user": fiber.Map{
			"id":      u.ID,
			"name":    u.Name,
			"surname": u.Surname,
			"email":   u.Email,
			"role":    u.Role,
		},
	})
}

// ListUsers godoc
// GET /backend_rootrevolution/api/users
// Requires admin
func (h *AuthHandler) ListUsers(c *fiber.Ctx) error {
	users, err := h.userSvc.ListUsers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve users",
		})
	}

	return c.JSON(fiber.Map{
		"data":  users,
		"total": len(users),
	})
}

// GetUser godoc
// GET /backend_rootrevolution/api/users/:id
func (h *AuthHandler) GetUser(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	u, err := h.userSvc.GetUser(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve user",
		})
	}
	if u == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.JSON(fiber.Map{"data": u})
}

// UpdateUser godoc
// PUT /backend_rootrevolution/api/users/:id
// Requires admin
func (h *AuthHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req appuser.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	requesterRole, _ := c.Locals("userRole").(user.Role)

	u, err := h.userSvc.UpdateUser(id, req, requesterRole)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "User updated successfully",
		"data":    u,
	})
}

// DeleteUser godoc
// DELETE /backend_rootrevolution/api/users/:id
// Requires admin
func (h *AuthHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	requesterRole, _ := c.Locals("userRole").(user.Role)

	if err := h.userSvc.DeleteUser(id, requesterRole); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{"message": "User deleted successfully"})
}

// Me godoc
// GET /backend_rootrevolution/api/auth/me
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"id":      c.Locals("userID"),
		"email":   c.Locals("userEmail"),
		"role":    c.Locals("userRole"),
		"name":    c.Locals("userName"),
		"surname": c.Locals("userSurname"),
	})
}
