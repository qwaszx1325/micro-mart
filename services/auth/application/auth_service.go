package application

import (
	"context"
	mmerror "micro-mart/pkg/mm_error"
	"micro-mart/pkg/mmotel"
	"micro-mart/services/auth/domain/service/service_impl"
	"time"

	"micro-mart/pkg/pb/gen/auth"
)

type AuthService struct {
	auth.UnimplementedAuthServiceServer
	authLogic *service_impl.AuthService
}

func NewAuthService(authLogic *service_impl.AuthService) *AuthService {
	return &AuthService{
		authLogic: authLogic,
	}
}

// 產生 token
func (s *AuthService) GenerateTokens(ctx context.Context, req *auth.GenerateTokenRequest) (*auth.AuthTokenResponse, error) {
	expTime := time.Now().Add(1 * time.Hour)
	accessToken, err := s.authLogic.GenerateAccessToken(ctx, req.UserId, req.Username, req.Email, req.Role, expTime)
	if err != nil {
		mmErr := mmerror.New(mmerror.InternalServerError, "generate access token fail")
		mmotel.Error(ctx, "generate access token fail", mmotel.NewField("err", err))
		return nil, mmErr
	}
	refreshExp := time.Now().Add(30 * 24 * time.Hour)
	refreshToken, err := s.authLogic.GenerateRefreshToken(ctx, req.UserId, req.Username, req.Email, req.Role, refreshExp)

	if err != nil {
		mmErr := mmerror.New(mmerror.InternalServerError, "generate refresh token fail")
		mmotel.Error(ctx, "generate refresh token fail", mmotel.NewField("err", err))
		return nil, mmErr
	}

	return &auth.AuthTokenResponse{
		Success:      true,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expTime.Unix(),
	}, nil
}

// 重新簽發 access token
func (s *AuthService) RefreshToken(ctx context.Context, req *auth.RefreshTokenRequest) (*auth.AuthTokenResponse, error) {
	accessToken, err := s.authLogic.RefreshAccessToken(ctx, req.RefreshToken)
	if err != nil {
		mmErr := mmerror.New(mmerror.InternalServerError, "refresh access token fail")
		mmotel.Error(ctx, "refresh access token fail", mmotel.NewField("err", err))
		return nil, mmErr
	}

	return &auth.AuthTokenResponse{
		Success:      true,
		AccessToken:  accessToken,
		RefreshToken: req.RefreshToken, // 保持原本 refresh token
		ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
	}, nil
}

// 解析 access token 拿 user profile
func (s *AuthService) GetUserProfile(ctx context.Context, req *auth.AccessTokenRequest) (*auth.GetUserProfileResponse, error) {
	claims, err := s.authLogic.ValidateJWT(ctx, req.AccessToken)
	if err != nil {
		mmErr := mmerror.New(mmerror.Unauthorized, "invalid access token")
		mmotel.Error(ctx, "invalid access token", mmotel.NewField("err", err))
		return nil, mmErr
	}
	return &auth.GetUserProfileResponse{
		UserId:   claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
		Role:     claims.Role,
	}, nil
}
