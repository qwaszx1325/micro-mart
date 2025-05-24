package application

import (
	"context"
	"micro-mart/pkg/db"
	mmerror "micro-mart/pkg/mm_error"
	"micro-mart/pkg/mmotel"
	"micro-mart/pkg/pb/gen/user"
	"micro-mart/pkg/utils"
	"micro-mart/services/user/domain/aggregate"
	"micro-mart/services/user/domain/entity"
	"micro-mart/services/user/domain/service"
	"micro-mart/services/user/domain/token_helper"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (s *UserService) Register(ctx context.Context, req *user.RegisterRequest) (*user.LoginResponse, error) {
	ctx, span := mmotel.StartSpan(ctx, "ApplicationUserService.Register")
	defer span.End()

	// Validate request
	if req == nil {
		mmotel.Error(ctx, "Register request is nil")
		return nil, status.Error(codes.InvalidArgument, "Register request cannot be nil")
	}

	if req.GetEmail() == "" || req.GetUsername() == "" || req.GetPassword() == "" {
		mmotel.Error(ctx, "Invalid register request",
			mmotel.NewField("email", req.GetEmail()),
			mmotel.NewField("username", req.GetUsername()))
		return nil, status.Error(codes.InvalidArgument, "Email, username and password are required")
	}

	hashPassword, _ := utils.HashPassword(req.GetPassword())

	// Create user profile
	profile := entity.Profile{
		Email:    req.GetEmail(),
		Name:     req.GetUsername(),
		Password: hashPassword,
	}

	u := &aggregate.User{
		Profile: profile,
	}

	// Begin transaction
	ctx, err := s.db.Begin(ctx)
	if err != nil {
		mmotel.Error(ctx, "Failed to begin transaction", mmotel.NewField("error", err))
		return nil, status.Error(codes.Internal, "Failed to process registration")
	}

	// Register user
	userProfile, mmErr := s.userService.Register(ctx, u)
	if mmErr != nil {
		// Rollback transaction
		_, rollbackErr := s.db.Rollback(ctx)
		if rollbackErr != nil {
			mmotel.Error(ctx, "Failed to rollback transaction",
				mmotel.NewField("rollbackError", rollbackErr),
				mmotel.NewField("originalError", mmErr))
		}

		// Map domain error to gRPC error
		code := codes.Internal
		switch mmErr.Code() {
		case mmerror.InvalidArgument:
			code = codes.InvalidArgument
		case mmerror.ResourceIsExist:
			code = codes.AlreadyExists
		}

		return nil, status.Error(code, mmErr.Message())
	}

	// Commit the transaction
	_, commitErr := s.db.Commit(ctx)
	if commitErr != nil {
		mmotel.Error(ctx, "Failed to commit transaction", mmotel.NewField("error", commitErr))
		return nil, status.Error(codes.Internal, "Failed to complete registration")
	}

	// Access Token：1小時有效
	accessTokenExpirationTime := time.Now().Add(1 * time.Hour)

	// Refresh Token：30天有效
	refreshTokenExpirationTime := time.Now().Add(30 * 24 * time.Hour)

	accessToken, err := token_helper.GenerateAccessToken(ctx, userProfile.ID, userProfile.Profile.Name, userProfile.Profile.Email, "測試用", accessTokenExpirationTime)
	if err != nil {
		mmotel.Error(ctx, "Failed to generate access token", mmotel.NewField("error", err))
		return nil, status.Error(codes.Internal, "Failed to generate access token")
	}

	refreshToken, err := token_helper.GenerateRefreshToken(ctx, userProfile.ID, userProfile.Profile.Name, userProfile.Profile.Email, "測試用", refreshTokenExpirationTime)
	if err != nil {
		mmotel.Error(ctx, "Failed to generate refresh token", mmotel.NewField("error", err))
		return nil, status.Error(codes.Internal, "Failed to generate refresh token")
	}

	expiresAt := int64(3600) // 1 hour in seconds

	return &user.LoginResponse{
		Message:      "Registration successful",
		Success:      true,
		UserId:       userProfile.ID.String(),
		Username:     userProfile.Profile.Name,
		Email:        userProfile.Profile.Email,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		// 暫時還沒有搞role所以先隨便設定
		Role: "user",
	}, nil
}

func (s *UserService) Login(ctx context.Context, req *user.LoginRequest) (*user.LoginResponse, error) {
	//ctx, span := mmotel.StartSpan(ctx, "ApplicationUserService.Login")
	//defer span.End()
	//
	//// Validate request
	//if req == nil {
	//	mmotel.Error(ctx, "Login request is nil")
	//	return nil, status.Error(codes.InvalidArgument, "Login request cannot be nil")
	//}
	//
	//if req.GetUsername() == "" || req.GetPassword() == "" {
	//	mmotel.Error(ctx, "Invalid login request",
	//		mmotel.NewField("username", req.GetUsername()))
	//	return nil, status.Error(codes.InvalidArgument, "Username and password are required")
	//}
	//
	//// Get user by username
	//user, mmErr := s.userService.GetUserByUserName(ctx, req.GetUsername())
	//if mmErr != nil {
	//	// Map domain error to gRPC error
	//	code := codes.Internal
	//	switch mmErr.Code() {
	//	case mmerror.InvalidArgument:
	//		code = codes.InvalidArgument
	//	case mmerror.ResourceNotFound:
	//		// Don't expose that the user doesn't exist for security reasons
	//		return nil, status.Error(codes.Unauthenticated, "Invalid username or password")
	//	}
	//
	//	mmotel.Error(ctx, "Failed to get user",
	//		mmotel.NewField("username", req.GetUsername()),
	//		mmotel.NewField("error", mmErr))
	//	return nil, status.Error(code, "Login failed")
	//}
	//
	//// Verify password (in a real app, you would use a secure password comparison)
	//if user.Profile.Password != req.GetPassword() {
	//	mmotel.Error(ctx, "Invalid password",
	//		mmotel.NewField("username", req.GetUsername()))
	//	return nil, status.Error(codes.Unauthenticated, "Invalid username or password")
	//}
	//
	//// Generate tokens (this is a placeholder - real implementation would generate actual tokens)
	//accessToken := "jwt-token-would-be-generated-here"
	//refreshToken := "refresh-token-would-be-generated-here"
	//expiresAt := int64(3600) // 1 hour in seconds
	//
	//return &user.AuthResponse{
	//	Success:       true,
	//	Message:       "Login successful",
	//	AccessToken:   accessToken,
	//	RefreshToken:  refreshToken,
	//	ExpiresAt:     expiresAt,
	//}, nil
	return nil, nil
}
