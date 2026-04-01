package user

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Surname   string    `json:"surname"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Role      Role      `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func NewUser(name, surname, email, hashedPassword string, role Role) *User {
	return &User{
		ID:        uuid.New(),
		Name:      name,
		Surname:   surname,
		Email:     email,
		Password:  hashedPassword,
		Role:      role,
		Status:    "active",
		CreatedAt: time.Now(),
	}
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}
