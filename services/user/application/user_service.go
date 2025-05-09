package application

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"micro-mart/pkg/db"
	"micro-mart/pkg/mmotel"
	"micro-mart/pkg/pb/gen/user"
	"micro-mart/services/user/domain/aggregate"
	"micro-mart/services/user/domain/entity"
	"micro-mart/services/user/domain/service"
)

type UserService struct {
	user.UserServiceServer
	userService *service.UserService
	db          db.Database
}

var _ user.UserServiceServer = (*UserService)(nil)

func NewUserService(userService *service.UserService, db db.Database) *UserService {
	return &UserService{
		userService: userService,
		db:          db,
	}
}

func (s *UserService) Register(ctx context.Context, req *user.RegisterRequest) (*user.AuthResponse, error) {
	ctx, span := mmotel.StartTrace(ctx)
	defer span.End()

	// 紀錄錯誤到追蹤系統（Jaeger）
	mmotel.Error(ctx, "模擬錯誤")
	return nil, status.Error(codes.Internal, "模擬錯誤")

	profile := entity.Profile{}

	profile.Email = req.GetEmail()
	profile.Name = req.GetUsername()
	profile.Password = req.GetPassword()
	u := &aggregate.User{
		Profile: profile,
	}

	ctx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	_, err = s.userService.Register(ctx, u)
	if err != nil {
		_, rollbackErr := s.db.Rollback(ctx)
		if rollbackErr != nil {
			mmotel.Error(ctx, rollbackErr.Error())
			err = rollbackErr
		}
		return nil, err
	}

	// Commit the transaction
	_, commitErr := s.db.Commit(ctx)
	if commitErr != nil {
		mmotel.Error(ctx, commitErr.Error())
		return nil, commitErr
	}

	return &user.AuthResponse{
		AccessToken:  "aaa",
		RefreshToken: "aaaa",
		Message:      "aaa",
		Success:      false,
		ExpiresAt:    123,
	}, nil
}

func (s *UserService) Login(ctx context.Context, req *user.LoginRequest) (*user.AuthResponse, error) {

	return nil, nil
}
func (s *UserService) RefreshToken(ctx context.Context, req *user.RefreshTokenRequest) (*user.AuthResponse, error) {

	return nil, nil
}
