package service

import (
	"context"
	"fmt"
	mmerror "micro-mart/pkg/mm_error"
	"micro-mart/pkg/mmotel"
	"micro-mart/services/user/domain/aggregate"
	"micro-mart/services/user/domain/repository"
	"strings"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) Register(ctx context.Context, user *aggregate.User) (*aggregate.User, *mmerror.MmError) {
	ctx, span := mmotel.StartSpan(ctx, "UserService.Register")
	defer span.End()

	// Call repository to register user
	registeredUser, mmErr := s.userRepo.RegisterUser(ctx, user)
	if mmErr != nil {
		return nil, mmerror.LogAndReturnError(
			ctx,
			mmErr.Code(),
			fmt.Sprintf("Failed to register user: %s", mmErr.Message()),
			mmErr,
			"User registration failed",
			mmotel.NewField("email", user.Profile.Email),
			mmotel.NewField("username", user.Profile.Name),
		)
	}

	return registeredUser, nil
}

func (s *UserService) GetUserByUserName(ctx context.Context, userName string) (*aggregate.User, *mmerror.MmError) {
	ctx, span := mmotel.StartSpan(ctx, "UserService.GetUserByUserName")
	defer span.End()

	// Validate username
	userName = strings.TrimSpace(userName)
	if userName == "" {
		return nil, mmerror.LogAndReturnError(
			ctx,
			mmerror.InvalidArgument,
			"Username cannot be empty",
			nil,
			"Empty username provided",
		)
	}

	// Call repository to get user
	user, mmErr := s.userRepo.GetUserByUserName(ctx, userName)
	if mmErr != nil {
		return nil, mmerror.LogAndReturnError(
			ctx,
			mmErr.Code(),
			fmt.Sprintf("Failed to get user by username: %s", mmErr.Message()),
			mmErr,
			"Failed to retrieve user",
			mmotel.NewField("username", userName),
		)
	}

	// Check if user was found
	if user == nil {
		return nil, mmerror.LogAndReturnError(
			ctx,
			mmerror.ResourceNotFound,
			fmt.Sprintf("User with username '%s' not found", userName),
			nil,
			"User not found",
			mmotel.NewField("username", userName),
		)
	}

	return user, nil
}
