package user

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"rootrevolution-api/config"
	"rootrevolution-api/internal/domain/user"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Name     string    `json:"name"`
	Surname  string    `json:"surname"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
	Role     user.Role `json:"role"`
}

type UpdateUserRequest struct {
	Name    *string    `json:"name"`
	Surname *string    `json:"surname"`
	Role    *user.Role `json:"role"`
	Status  *string    `json:"status"`
}

type Claims struct {
	UserID  string    `json:"user_id"`
	Email   string    `json:"email"`
	Role    user.Role `json:"role"`
	Name    string    `json:"name"`
	Surname string    `json:"surname"`
	jwt.RegisteredClaims
}

type Service struct {
	userRepo user.Repository
	cfg      *config.Config
}

func NewService(userRepo user.Repository, cfg *config.Config) *Service {
	return &Service{userRepo: userRepo, cfg: cfg}
}

func (s *Service) EnsureDefaultAdmin() error {
	exists, err := s.userRepo.Exists("biangacila@gmail.com")
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte("Nathan010309*"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := user.NewUser("Merveilleux", "Biangacila", "biangacila@gmail.com", string(hashed), user.RoleAdmin)
	return s.userRepo.Save(admin)
}

func (s *Service) Login(req LoginRequest) (string, *user.User, error) {
	u, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return "", nil, err
	}
	if u == nil {
		return "", nil, fmt.Errorf("invalid credentials")
	}
	if u.Status != "active" {
		return "", nil, fmt.Errorf("account is not active")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		return "", nil, fmt.Errorf("invalid credentials")
	}

	token, err := s.generateToken(u)
	if err != nil {
		return "", nil, fmt.Errorf("generating token: %w", err)
	}

	return token, u, nil
}

func (s *Service) Register(req RegisterRequest, createdByRole user.Role) (*user.User, error) {
	if createdByRole != user.RoleAdmin {
		return nil, fmt.Errorf("only admins can register new users")
	}

	exists, err := s.userRepo.Exists(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	role := req.Role
	if role == "" {
		role = user.RoleEditor
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	u := user.NewUser(req.Name, req.Surname, req.Email, string(hashed), role)
	if err := s.userRepo.Save(u); err != nil {
		return nil, fmt.Errorf("saving user: %w", err)
	}

	return u, nil
}

func (s *Service) ListUsers() ([]user.User, error) {
	return s.userRepo.FindAll()
}

func (s *Service) GetUser(id uuid.UUID) (*user.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *Service) UpdateUser(id uuid.UUID, req UpdateUserRequest, requesterRole user.Role) (*user.User, error) {
	if requesterRole != user.RoleAdmin {
		return nil, fmt.Errorf("only admins can update users")
	}

	u, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, fmt.Errorf("user not found")
	}

	if req.Name != nil {
		u.Name = *req.Name
	}
	if req.Surname != nil {
		u.Surname = *req.Surname
	}
	if req.Role != nil {
		u.Role = *req.Role
	}
	if req.Status != nil {
		u.Status = *req.Status
	}

	if err := s.userRepo.Update(u); err != nil {
		return nil, fmt.Errorf("updating user: %w", err)
	}

	return u, nil
}

func (s *Service) DeleteUser(id uuid.UUID, requesterRole user.Role) error {
	if requesterRole != user.RoleAdmin {
		return fmt.Errorf("only admins can delete users")
	}
	return s.userRepo.Delete(id)
}

func (s *Service) generateToken(u *user.User) (string, error) {
	claims := Claims{
		UserID:  u.ID.String(),
		Email:   u.Email,
		Role:    u.Role,
		Name:    u.Name,
		Surname: u.Surname,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   u.Email,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.Secret))
}

func (s *Service) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.cfg.JWT.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
