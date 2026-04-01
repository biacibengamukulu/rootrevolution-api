package user

import "github.com/google/uuid"

type Repository interface {
	FindByEmail(email string) (*User, error)
	FindByID(id uuid.UUID) (*User, error)
	FindAll() ([]User, error)
	Save(u *User) error
	Update(u *User) error
	Delete(id uuid.UUID) error
	Exists(email string) (bool, error)
}
