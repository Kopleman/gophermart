package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Kopleman/gophermart/internal/common/dto"
	"github.com/Kopleman/gophermart/internal/common/log"
	"github.com/Kopleman/gophermart/internal/config"
	"github.com/Kopleman/gophermart/internal/pgxstore"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepo interface {
	CreateNewUser(ctx context.Context, createDto *dto.CreateUserDTO) (*pgxstore.User, error)
	GetUser(ctx context.Context, login string) (*pgxstore.User, error)
	GetUserWithdrawals(ctx context.Context, userID uuid.UUID) ([]*pgxstore.Transaction, error)
}

type UserService struct {
	logger   log.Logger
	cfg      *config.Config
	userRepo UserRepo
}

func NewUserService(
	logger log.Logger,
	cfg *config.Config,
	userRepo UserRepo,
) *UserService {
	return &UserService{
		logger,
		cfg,
		userRepo,
	}
}

func (s *UserService) CreateUser(ctx context.Context, createReqDto *dto.UserCredentialsDTO) error {
	existed, err := s.userRepo.GetUser(ctx, createReqDto.Login)
	if err != nil && !errors.Is(err, pgxstore.ErrNotFound) {
		return fmt.Errorf("userService.createUser.getUser: %w", err)
	}
	if existed != nil {
		return ErrAlreadyExists
	}

	hashedPassword, hashError := s.hashPassword(createReqDto.Password)
	if hashError != nil {
		return fmt.Errorf("userService.createUser.hashPassword.: %w", err)
	}

	createDto := &dto.CreateUserDTO{
		Login:        createReqDto.Login,
		PasswordHash: hashedPassword,
	}

	if _, createError := s.userRepo.CreateNewUser(ctx, createDto); createError != nil {
		return fmt.Errorf("userService.createUser.userRepo.CreateNewUser: %w", createError)
	}

	return nil
}

func (s *UserService) AuthorizeUser(ctx context.Context, loginDto *dto.UserCredentialsDTO) (string, error) {
	user, err := s.userRepo.GetUser(ctx, loginDto.Login)
	if err != nil {
		if errors.Is(err, pgxstore.ErrNotFound) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("userService.authorizeUser.getUser: %w", err)
	}
	passwordOk := s.verifyPassword(loginDto.Password, user.PasswordHash)
	if !passwordOk {
		return "", ErrInvalidArguments
	}
	newToken, tokenErr := s.generateToken(user.ID.String())
	if tokenErr != nil {
		return "", fmt.Errorf("userService.authorizeUser.generateToken: %w", tokenErr)
	}

	return newToken, nil
}

func (s *UserService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("cant hash password: %w", err)
	}
	return string(bytes), err
}

func (s *UserService) verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *UserService) generateToken(id string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": id,
	})

	t, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("cant generate token: %w", err)
	}

	return t, nil
}

func (s *UserService) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]*dto.WithdrawalItemDTO, error) {
	withdrawalTxs, err := s.userRepo.GetUserWithdrawals(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("userService.GetWithdrawals: %w", err)
	}
	dtos := make([]*dto.WithdrawalItemDTO, len(withdrawalTxs))
	for i, withdrawal := range withdrawalTxs {
		dtos[i] = withdrawal.ToWithdrawalItemDTO()
	}
	return dtos, nil
}
