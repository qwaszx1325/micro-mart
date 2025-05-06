package service

import (
	"context"
	mmerror "micro-mart/pkg/mm_error"
	"micro-mart/pkg/mmotel"
	"micro-mart/services/user/domain/aggregate"
	"micro-mart/services/user/domain/repository"
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
	ctx, span := mmotel.StartTrace(ctx)
	defer span.End()

	return s.userRepo.RegisterUser(ctx, user)
}

func (s *UserService) GetUserByUserName(ctx context.Context, userName string) (*aggregate.User, *mmerror.MmError) {
	ctx, span := mmotel.StartTrace(ctx)
	defer span.End()

	return s.userRepo.GetUserByUserName(ctx, userName)
}
