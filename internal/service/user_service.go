package service

import (
	"context"
	"fmt"

	"go-sqlite-api/internal/domain"
)

type UserService interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)
	GetAllUsers(ctx context.Context) ([]domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	DeleteUser(ctx context.Context, id int64) error
}

type userService struct {
	userRepo   domain.UserRepository
	notifier   domain.NotificationService
}

func NewUserService(userRepo domain.UserRepository, notifier domain.NotificationService) UserService {
	return &userService{
		userRepo: userRepo,
		notifier: notifier,
	}
}

func (s *userService) CreateUser(ctx context.Context, user *domain.User) error {
	if err := s.userRepo.Create(ctx, user); err != nil {
		return err
	}

	// Trigger Gotify Notification
	go func() {
		msg := fmt.Sprintf("New user registered: %s (%s)", user.Name, user.Email)
		_ = s.notifier.Send(context.Background(), "User Registration", msg, 5)
	}()

	return nil
}

func (s *userService) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *userService) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	return s.userRepo.GetAll(ctx)
}

func (s *userService) UpdateUser(ctx context.Context, user *domain.User) error {
	return s.userRepo.Update(ctx, user)
}

func (s *userService) DeleteUser(ctx context.Context, id int64) error {
	return s.userRepo.Delete(ctx, id)
}
