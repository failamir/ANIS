package services_test

import (
	"log"
	"testing"

	"auth-service/internal/config"
	"auth-service/internal/models"
	"auth-service/internal/repository/mocks"
	"auth-service/internal/services"
	"auth-service/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister_Success(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.MockUserRepository)
	mockAuthRepo := new(mocks.MockAuthRepository)
	cfg := &config.Config{
		JWTSecret:     "secret",
		RefreshSecret: "refresh",
	}

	service := services.NewAuthService(mockUserRepo, mockAuthRepo, cfg)

	// Expectations
	email := "test@example.com"
	password := "password123"
	name := "Test User"

	// Mock FindByEmail to return nil (user doesn't exist)
	mockUserRepo.On("FindByEmail", email).Return(nil, nil)

	// Mock CreateUser to return nil (success)
	mockUserRepo.On("CreateUser", mock.AnythingOfType("*models.User")).Return(nil)

	// Execute
	err := service.Register(name, email, password)

	// Assert
	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.MockUserRepository)
	mockAuthRepo := new(mocks.MockAuthRepository)
	cfg := &config.Config{
		JWTSecret:     "secret",
		RefreshSecret: "refresh",
	}

	service := services.NewAuthService(mockUserRepo, mockAuthRepo, cfg)

	// Prepare data
	email := "test@example.com"
	password := "password123"
	hashedPassword, _ := utils.HashPassword(password)

	user := &models.User{
		ID:       1,
		Email:    email,
		Password: hashedPassword,
		Name:     "Test User",
	}

	// Expectations
	mockUserRepo.On("FindByEmail", email).Return(user, nil)

	// Mock AuthRepo CreateAuth
	// We use mock.Anything for UUIDs because they are random
	mockAuthRepo.On("CreateAuth", user.ID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Execute
	token, err := service.Login(email, password)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.NotEmpty(t, token.AccessToken)

	mockUserRepo.AssertExpectations(t)
	mockAuthRepo.AssertExpectations(t)
}

func TestLogin_InvalidPassword(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.MockUserRepository)
	mockAuthRepo := new(mocks.MockAuthRepository)
	cfg := &config.Config{}

	service := services.NewAuthService(mockUserRepo, mockAuthRepo, cfg)

	// Data
	email := "test@example.com"
	realPassword := "password123"
	wrongPassword := "wrongpass"
	hashedPassword, _ := utils.HashPassword(realPassword)

	user := &models.User{
		ID:       1,
		Email:    email,
		Password: hashedPassword,
	}

	// Expectations
	mockUserRepo.On("FindByEmail", email).Return(user, nil)

	// Execute
	token, err := service.Login(email, wrongPassword)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, token)
	assert.Equal(t, "invalid credentials", err.Error())
}

func init() {
	// Quiet logs during test
	log.SetFlags(0)
}
