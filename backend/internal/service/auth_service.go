package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/apperr"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/config"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/model"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/pagination"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/repository"
)

type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

type RegisterRequest struct {
	Email     string  `json:"email"`
	Password  string  `json:"password"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Role      *string `json:"role,omitempty"`
}

type AuthResponse struct {
	User        *model.User `json:"user"`
	AccessToken string      `json:"access_token"`
	TokenType   string      `json:"token_type"`
	ExpiresIn   int64       `json:"expires_in"`
}

type AuthService interface {
	Login(ctx context.Context, email, password string) (*AuthResponse, error)
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*model.User, error)
	ValidateToken(tokenStr string) (*JWTClaims, error)
	ListUsers(ctx context.Context, params pagination.Params) ([]model.User, int64, error)
}

type authService struct {
	userRepo repository.UserRepository
	jwtCfg   config.JWT
}

func NewAuthService(userRepo repository.UserRepository, jwtCfg config.JWT) AuthService {
	return &authService{
		userRepo: userRepo,
		jwtCfg:   jwtCfg,
	}
}

func (s *authService) Login(ctx context.Context, email, password string) (*AuthResponse, error) {
	email = strings.TrimSpace(email)
	if email == "" || password == "" {
		return nil, apperr.Validation(apperr.Fields{
			"email":    []string{"email is required"},
			"password": []string{"password is required"},
		})
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if user == nil || !user.CheckPassword(password) {
		return nil, apperr.Unauthorized("Invalid email or password")
	}

	if !user.IsActive {
		return nil, apperr.Forbidden("Account is disabled")
	}

	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		// Non-fatal warning log internally; proceed with login
	}

	return s.generateAuthResponse(user)
}

func (s *authService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	email := strings.TrimSpace(req.Email)
	firstName := strings.TrimSpace(req.FirstName)
	lastName := strings.TrimSpace(req.LastName)

	fields := apperr.Fields{}
	if email == "" {
		fields["email"] = []string{"email is required"}
	}
	if req.Password == "" || len(req.Password) < 6 {
		fields["password"] = []string{"password must be at least 6 characters"}
	}
	if firstName == "" {
		fields["first_name"] = []string{"first_name is required"}
	}
	if lastName == "" {
		fields["last_name"] = []string{"last_name is required"}
	}
	if len(fields) > 0 {
		return nil, apperr.Validation(fields)
	}

	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if existing != nil {
		return nil, apperr.Conflict("Email %q is already registered", email)
	}

	role := model.RoleUser
	if req.Role != nil && *req.Role != "" {
		switch *req.Role {
		case model.RoleAdmin, model.RoleManager, model.RoleUser:
			role = *req.Role
		default:
			return nil, apperr.Validation(apperr.Fields{"role": []string{"invalid role"}})
		}
	}

	user := &model.User{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
		IsActive:  true,
	}

	if err := user.SetPassword(req.Password); err != nil {
		return nil, apperr.Internal(err)
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, apperr.Internal(err)
	}

	return s.generateAuthResponse(user)
}

func (s *authService) GetProfile(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if user == nil {
		return nil, apperr.NotFound("user")
	}
	return user, nil
}

func (s *authService) ValidateToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtCfg.Secret), nil
	})

	if err != nil || !token.Valid {
		return nil, apperr.Unauthorized("Invalid or expired token")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, apperr.Unauthorized("Invalid token claims")
	}

	return claims, nil
}

func (s *authService) ListUsers(ctx context.Context, params pagination.Params) ([]model.User, int64, error) {
	users, total, err := s.userRepo.List(ctx, params)
	if err != nil {
		return nil, 0, apperr.Internal(err)
	}
	return users, total, nil
}

func (s *authService) generateAuthResponse(user *model.User) (*AuthResponse, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(s.jwtCfg.AccessTokenTTL)

	claims := &JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(s.jwtCfg.Secret))
	if err != nil {
		return nil, apperr.Internal(err)
	}

	return &AuthResponse{
		User:        user,
		AccessToken: tokenStr,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.jwtCfg.AccessTokenTTL.Seconds()),
	}, nil
}
