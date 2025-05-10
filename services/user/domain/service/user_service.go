package service

import (
	"context"
	"fmt"
	mmerror "micro-mart/pkg/mm_error"
	"micro-mart/pkg/mmotel"
	"micro-mart/services/user/domain/aggregate"
	"micro-mart/services/user/domain/repository"
	"regexp"
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

// validateEmail checks if the email is valid
func validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	// Simple email validation using regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format: %s", email)
	}

	return nil
}

// validateUsername checks if the username is valid
func validateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}

	return nil
}

// validatePassword checks if the password is valid
func validatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	return nil
}

// validateUser validates all user fields
func validateUser(user *aggregate.User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	if err := validateEmail(user.Profile.Email); err != nil {
		return err
	}

	if err := validateUsername(user.Profile.Name); err != nil {
		return err
	}

	if err := validatePassword(user.Profile.Password); err != nil {
		return err
	}

	return nil
}

func (s *UserService) Register(ctx context.Context, user *aggregate.User) (*aggregate.User, *mmerror.MmError) {
	ctx, span := mmotel.StartSpan(ctx, "UserService.Register")
	defer span.End()

	// Validate user input
	if err := validateUser(user); err != nil {
		return nil, mmerror.LogAndReturnError(
			ctx,
			mmerror.InvalidArgument,
			fmt.Sprintf("Invalid user data: %v", err),
			err,
			"User validation failed",
			mmotel.NewField("email", user.Profile.Email),
			mmotel.NewField("username", user.Profile.Name),
		)
	}

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
