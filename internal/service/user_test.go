package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Kopleman/gophermart/internal/common/dto"
	"github.com/Kopleman/gophermart/internal/common/log"
	"github.com/Kopleman/gophermart/internal/config"
	"github.com/Kopleman/gophermart/internal/pgxstore"
	"github.com/Kopleman/gophermart/internal/service/mocks"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestUserService_CreateUser(t *testing.T) {
	cfg := &config.Config{JWTSecret: "secret"}
	logger := log.New()
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		createDto := &dto.UserCredentialsDTO{
			Login:    "testuser",
			Password: "password",
		}
		repo := new(mocks.UserRepo)
		repo.On("GetUser", ctx, createDto.Login).Return(nil, pgxstore.ErrNotFound)
		repo.On("CreateNewUser", ctx, mock.Anything).Return(&pgxstore.User{ID: uuid.New()}, nil)

		service := NewUserService(log.MockLogger{}, cfg, repo)
		err := service.CreateUser(ctx, createDto)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("user already exists", func(t *testing.T) {
		createDto := &dto.UserCredentialsDTO{
			Login:    "existinguser",
			Password: "password",
		}
		repo := new(mocks.UserRepo)
		repo.On("GetUser", ctx, createDto.Login).Return(&pgxstore.User{ID: uuid.New()}, nil)

		service := NewUserService(logger, cfg, repo)
		err := service.CreateUser(ctx, createDto)

		assert.ErrorIs(t, err, ErrAlreadyExists)
		repo.AssertExpectations(t)
	})

	t.Run("repository error on user get", func(t *testing.T) {
		createDto := &dto.UserCredentialsDTO{
			Login:    "testuser",
			Password: "password",
		}
		expectedErr := errors.New("database error")
		repo := new(mocks.UserRepo)
		repo.On("GetUser", ctx, createDto.Login).Return(nil, expectedErr)

		service := NewUserService(logger, cfg, repo)
		err := service.CreateUser(ctx, createDto)

		assert.ErrorIs(t, err, expectedErr)
		repo.AssertExpectations(t)
	})

	t.Run("repository error on user creation", func(t *testing.T) {
		createDto := &dto.UserCredentialsDTO{
			Login:    "erroruser",
			Password: "password",
		}
		expectedErr := errors.New("database error")
		repo := new(mocks.UserRepo)
		repo.On("GetUser", ctx, createDto.Login).Return(nil, pgxstore.ErrNotFound)
		repo.On("CreateNewUser", ctx, mock.Anything).Return(nil, expectedErr)

		service := NewUserService(logger, cfg, repo)
		err := service.CreateUser(ctx, createDto)

		assert.ErrorIs(t, err, expectedErr)
		repo.AssertExpectations(t)
	})
}

func TestUserService_AuthorizeUser(t *testing.T) {
	cfg := &config.Config{JWTSecret: "secret"}
	logger := log.New()
	ctx := context.Background()
	userID := uuid.New()
	correctPassword := "password"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)

	t.Run("successful authorization", func(t *testing.T) {
		creds := &dto.UserCredentialsDTO{
			Login:    "validuser",
			Password: correctPassword,
		}
		repo := new(mocks.UserRepo)
		repo.On("GetUser", ctx, creds.Login).Return(&pgxstore.User{
			ID:           userID,
			PasswordHash: string(hashedPassword),
		}, nil)

		service := NewUserService(logger, cfg, repo)
		token, err := service.AuthorizeUser(ctx, creds)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		repo.AssertExpectations(t)
	})

	t.Run("invalid password", func(t *testing.T) {
		creds := &dto.UserCredentialsDTO{
			Login:    "validuser",
			Password: "wrongpassword",
		}
		repo := new(mocks.UserRepo)
		repo.On("GetUser", ctx, creds.Login).Return(&pgxstore.User{
			ID:           userID,
			PasswordHash: string(hashedPassword),
		}, nil)

		service := NewUserService(logger, cfg, repo)
		_, err := service.AuthorizeUser(ctx, creds)

		assert.ErrorIs(t, err, ErrInvalidArguments)
		repo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		creds := &dto.UserCredentialsDTO{
			Login:    "nonexistent",
			Password: "password",
		}
		repo := new(mocks.UserRepo)
		repo.On("GetUser", ctx, creds.Login).Return(nil, pgxstore.ErrNotFound)

		service := NewUserService(logger, cfg, repo)
		_, err := service.AuthorizeUser(ctx, creds)

		assert.ErrorIs(t, err, ErrNotFound)
		repo.AssertExpectations(t)
	})
}

func TestUserService_GetWithdrawals(t *testing.T) {
	cfg := &config.Config{}
	logger := log.New()
	ctx := context.Background()
	userID := uuid.New()

	t.Run("successful get withdrawals", func(t *testing.T) {
		repo := new(mocks.UserRepo)
		orderNumber := "49927398716"
		createdAt := time.Now()
		sum := decimal.NewFromFloat(100.0)
		sumAsFloat, _ := sum.Float64()
		expectedWithdrawals := []*dto.WithdrawalItemDTO{
			{
				Order:       orderNumber,
				ProcessedAt: createdAt.Format(time.RFC3339),
				Sum:         sumAsFloat,
			},
		}
		expected := []*pgxstore.Transaction{
			{
				ID:          uuid.New(),
				UserID:      userID,
				OrderNumber: &orderNumber,
				OldBalance:  sum,
				Change:      sum,
				NewBalance:  decimal.Zero,
				Type:        pgxstore.TransactionTypeWITHDRAW,
				CreatedAt:   createdAt,
			},
		}
		repo.On("GetUserWithdrawals", ctx, userID).Return(expected, nil)

		service := NewUserService(logger, cfg, repo)
		result, err := service.GetWithdrawals(ctx, userID)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, expectedWithdrawals, result)
		repo.AssertExpectations(t)
	})

	t.Run("no withdrawals found", func(t *testing.T) {
		repo := new(mocks.UserRepo)
		repo.On("GetUserWithdrawals", ctx, userID).Return([]*pgxstore.Transaction{}, nil)

		service := NewUserService(logger, cfg, repo)
		result, err := service.GetWithdrawals(ctx, userID)

		assert.NoError(t, err)
		assert.Empty(t, result)
		repo.AssertExpectations(t)
	})
}

func TestPasswordHashing(t *testing.T) {
	service := NewUserService(nil, nil, nil)
	password := "secretpassword"

	hash, err := service.hashPassword(password)
	assert.NoError(t, err)
	assert.True(t, service.verifyPassword(password, hash))
	assert.False(t, service.verifyPassword("wrongpassword", hash))
}

func TestTokenGeneration(t *testing.T) {
	cfg := &config.Config{JWTSecret: "testsecret"}
	service := NewUserService(nil, cfg, nil)
	userID := uuid.New().String()

	token, err := service.generateToken(userID)
	assert.NoError(t, err)

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})
	assert.NoError(t, err)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, userID, claims["userId"])
}
