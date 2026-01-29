package services

import (
	"errors"

	"auth-service/internal/config"
	"auth-service/internal/models"
	"auth-service/internal/repository"
	"auth-service/internal/utils"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	userRepo repository.UserRepository
	authRepo repository.AuthRepository
	cfg      *config.Config
}

func NewAuthService(userRepo repository.UserRepository, authRepo repository.AuthRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		authRepo: authRepo,
		cfg:      cfg,
	}
}

func (s *AuthService) Register(name, email, password string) error {
	existingUser, _ := s.userRepo.FindByEmail(email)
	if existingUser != nil && existingUser.ID != 0 {
		return errors.New("email already registered")
	}

	hashed, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	user := &models.User{
		Name:     name,
		Email:    email,
		Password: hashed,
	}

	return s.userRepo.CreateUser(user)
}

func (s *AuthService) Login(email, password string) (*utils.TokenDetails, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	match, err := utils.VerifyPassword(password, user.Password)
	if err != nil || !match {
		return nil, errors.New("invalid credentials")
	}

	td, err := utils.GenerateToken(user.ID, s.cfg)
	if err != nil {
		return nil, err
	}

	// Save token metadata to Redis via AuthRepo
	err = s.authRepo.CreateAuth(user.ID, td.AccessUuid, td.RefreshUuid, td.AtExpires, td.RtExpires)
	if err != nil {
		return nil, err
	}

	return td, nil
}

func (s *AuthService) Refresh(refreshToken string) (*utils.TokenDetails, error) {
	// Verify Token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.RefreshSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	refreshUuid, ok := claims["refresh_uuid"].(string)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	userIdFloat, ok := claims["user_id"].(float64)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	userId := uint(userIdFloat)

	// Check if token exists in Redis
	val, err := s.authRepo.FetchAuth(refreshUuid)
	if err != nil || val == "" {
		return nil, errors.New("token expired or revoked")
	}

	// Delete old metadata (Rotation)
	s.authRepo.DeleteAuth(refreshUuid)

	td, err := utils.GenerateToken(userId, s.cfg)
	if err != nil {
		return nil, err
	}

	// Register new token pair
	err = s.authRepo.CreateAuth(userId, td.AccessUuid, td.RefreshUuid, td.AtExpires, td.RtExpires)
	if err != nil {
		return nil, err
	}

	return td, nil
}
