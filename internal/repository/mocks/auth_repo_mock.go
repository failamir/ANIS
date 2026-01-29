package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) CreateAuth(userid uint, accessUuid, refreshUuid string, atExpires, rtExpires int64) error {
	args := m.Called(userid, accessUuid, refreshUuid, atExpires, rtExpires)
	return args.Error(0)
}

func (m *MockAuthRepository) FetchAuth(uuid string) (string, error) {
	args := m.Called(uuid)
	return args.String(0), args.Error(1)
}

func (m *MockAuthRepository) DeleteAuth(uuid string) error {
	args := m.Called(uuid)
	return args.Error(0)
}
