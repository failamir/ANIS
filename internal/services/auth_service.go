package services

import (
	"context"
	"errors"
	"time"

	"auth-service/internal/config"
	"auth-service/internal/models"
	"auth-service/internal/repository"
	"auth-service/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

type AuthService struct {
	repo  repository.UserRepository
	redis *redis.Client
	cfg   *config.Config
}

func NewAuthService(repo repository.UserRepository, rdb *redis.Client, cfg *config.Config) *AuthService {
	return &AuthService{
		repo:  repo,
		redis: rdb,
		cfg:   cfg,
	}
}

func (s *AuthService) Register(name, email, password string) error {
	existingUser, _ := s.repo.FindByEmail(email)
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

	return s.repo.CreateUser(user)
}

func (s *AuthService) Login(email, password string) (*utils.TokenDetails, error) {
	user, err := s.repo.FindByEmail(email)
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

	// Save token metadata to Redis
	err = s.CreateAuth(user.ID, td)
	if err != nil {
		return nil, err
	}

	return td, nil
}

func (s *AuthService) CreateAuth(userid uint, td *utils.TokenDetails) error {
	at := time.Unix(td.AtExpires, 0) // converting Unix to UTC
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()

	err := s.redis.Set(context.Background(), td.AccessUuid, string(userid), at.Sub(now)).Err()
	if err != nil {
		return err
	}
	err = s.redis.Set(context.Background(), td.RefreshUuid, string(userid), rt.Sub(now)).Err()
	if err != nil {
		return err
	}
	return nil
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

	// Check if token exists in Redis (if not, it might be revoked or expired)
	// In a real strict implementation, we would `Get` it.
	// The PRD mentions "Revoke", so checking Redis is crucial.
	ctx := context.Background()
	val, err := s.redis.Get(ctx, refreshUuid).Result()
	if err != nil {
		return nil, errors.New("token expired or revoked")
	}

	// Delete old metadata (Rotation)
	s.redis.Del(ctx, refreshUuid)

	// Issue new pair
	if val == "" { // Should not happen given Get error check, but safe guard
		return nil, errors.New("unauthorized")
	}

	td, err := utils.GenerateToken(userId, s.cfg)
	if err != nil {
		return nil, err
	}

	err = s.CreateAuth(userId, td)
	if err != nil {
		return nil, err
	}

	return td, nil
}
