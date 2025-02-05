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
	"golang.org/x/crypto/bcrypt"
)

type Store interface {
	CreateNewUser(ctx context.Context, createDto *dto.CreateUserDTO) (*pgxstore.User, error)
	GetUser(ctx context.Context, login string) (*pgxstore.User, error)
}

type Service struct {
	logger log.Logger
	config *config.Config
	store  Store
}

func NewUserService(
	logger log.Logger,
	config *config.Config,
	store Store,
) *Service {
	return &Service{
		logger,
		config,
		store,
	}
}

func (s *Service) CreateUser(ctx context.Context, createReqDto *dto.CreateUserRequestDTO) error {
	existed, err := s.store.GetUser(ctx, createReqDto.Login)
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

	if _, createError := s.store.CreateNewUser(ctx, createDto); createError != nil {
		return fmt.Errorf("userService.createUser.store.CreateNewUser: %w", createError)
	}

	return nil
}

func (s *Service) AuthorizeUser(ctx context.Context, loginDto *dto.UserLoginRequestDTO) (string, error) {
	user, err := s.store.GetUser(ctx, loginDto.Login)
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

func (s *Service) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("cant hash password: %w", err)
	}
	return string(bytes), err
}

func (s *Service) verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *Service) generateToken(id string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": id,
	})

	t, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", err
	}

	return t, nil
}

func (s *Service) VerifyToken(tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.JWTSecret), nil
	})
	if err != nil {
		return false, err
	}

	return token.Valid, nil
}
