package service

import (
	"Practice8/repository"
	"Practice8/service/mocks"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetUserByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	userService := NewUserService(mockRepo)

	user := &repository.User{ID: 1, Name: "John Doe"}

	mockRepo.EXPECT().GetUserByID(1).Return(user, nil)

	result, err := userService.GetUserByID(1)

	assert.NoError(t, err)
	assert.Equal(t, user, result)
}

func TestCreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	userService := NewUserService(mockRepo)

	user := &repository.User{ID: 2, Name: "Jane Doe"}

	mockRepo.EXPECT().CreateUser(user).Return(nil)

	err := userService.CreateUser(user)

	assert.NoError(t, err)
}

func TestRegisterUserTableDriven(t *testing.T) {
	tests := []struct {
		name         string
		user         *repository.User
		email        string
		existingUser *repository.User
		repoErr      error
		expectedErr  string
		setupMock    func(*mocks.MockUserRepository)
	}{
		{
			name:         "user already exists",
			user:         &repository.User{ID: 1, Name: "John"},
			email:        "john@example.com",
			existingUser: &repository.User{ID: 1, Name: "John"},
			repoErr:      nil,
			expectedErr:  "user with this email already exists",
			setupMock: func(m *mocks.MockUserRepository) {
				m.EXPECT().GetByEmail("john@example.com").Return(
					&repository.User{ID: 1, Name: "John"}, nil)
			},
		},
		{
			name:         "successful registration",
			user:         &repository.User{ID: 2, Name: "Jane"},
			email:        "jane@example.com",
			existingUser: nil,
			repoErr:      nil,
			expectedErr:  "",
			setupMock: func(m *mocks.MockUserRepository) {
				gomock.InOrder(
					m.EXPECT().GetByEmail("jane@example.com").Return(nil, nil),
					m.EXPECT().CreateUser(&repository.User{ID: 2, Name: "Jane"}).Return(nil),
				)
			},
		},
		{
			name:         "repository error on get",
			user:         &repository.User{ID: 3, Name: "Bob"},
			email:        "bob@example.com",
			existingUser: nil,
			repoErr:      fmt.Errorf("db error"),
			expectedErr:  "error getting user with this email",
			setupMock: func(m *mocks.MockUserRepository) {
				m.EXPECT().GetByEmail("bob@example.com").Return(nil, fmt.Errorf("db error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockUserRepository(ctrl)
			tt.setupMock(mockRepo)

			userService := NewUserService(mockRepo)
			err := userService.RegisterUser(tt.user, tt.email)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateUserNameTableDriven(t *testing.T) {
	tests := []struct {
		name         string
		userID       int
		newName      string
		existingUser *repository.User
		expectedErr  string
		setupMock    func(*mocks.MockUserRepository)
	}{
		{
			name:         "empty name",
			userID:       1,
			newName:      "",
			existingUser: nil,
			expectedErr:  "name cannot be empty",
			setupMock: func(m *mocks.MockUserRepository) {
			},
		},
		{
			name:         "user not found",
			userID:       99,
			newName:      "New Name",
			existingUser: nil,
			expectedErr:  "user not found",
			setupMock: func(m *mocks.MockUserRepository) {
				m.EXPECT().GetUserByID(99).Return(nil, fmt.Errorf("user not found"))
			},
		},
		{
			name:         "successful update",
			userID:       1,
			newName:      "Updated Name",
			existingUser: &repository.User{ID: 1, Name: "Old Name"},
			expectedErr:  "",
			setupMock: func(m *mocks.MockUserRepository) {
				gomock.InOrder(
					m.EXPECT().GetUserByID(1).Return(&repository.User{ID: 1, Name: "Old Name"}, nil),
					m.EXPECT().UpdateUser(&repository.User{ID: 1, Name: "Updated Name"}).Return(nil),
				)
			},
		},
		{
			name:         "update fails",
			userID:       1,
			newName:      "New Name",
			existingUser: &repository.User{ID: 1, Name: "Old Name"},
			expectedErr:  "database error",
			setupMock: func(m *mocks.MockUserRepository) {
				gomock.InOrder(
					m.EXPECT().GetUserByID(1).Return(&repository.User{ID: 1, Name: "Old Name"}, nil),
					m.EXPECT().UpdateUser(&repository.User{ID: 1, Name: "New Name"}).Return(fmt.Errorf("database error")),
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockUserRepository(ctrl)
			tt.setupMock(mockRepo)

			userService := NewUserService(mockRepo)
			err := userService.UpdateUserName(tt.userID, tt.newName)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteUserTableDriven(t *testing.T) {
	tests := []struct {
		name        string
		userID      int
		expectedErr string
		setupMock   func(*mocks.MockUserRepository)
	}{
		{
			name:        "cannot delete admin",
			userID:      1,
			expectedErr: "it is not allowed to delete admin user",
			setupMock: func(m *mocks.MockUserRepository) {
				// Не вызываем DeleteUser для админа
			},
		},
		{
			name:        "successful delete",
			userID:      2,
			expectedErr: "",
			setupMock: func(m *mocks.MockUserRepository) {
				m.EXPECT().DeleteUser(2).Return(nil)
			},
		},
		{
			name:        "delete fails",
			userID:      3,
			expectedErr: "database error",
			setupMock: func(m *mocks.MockUserRepository) {
				m.EXPECT().DeleteUser(3).Return(fmt.Errorf("database error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockUserRepository(ctrl)
			tt.setupMock(mockRepo)

			userService := NewUserService(mockRepo)
			err := userService.DeleteUser(tt.userID)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
