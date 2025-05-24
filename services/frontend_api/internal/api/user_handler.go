package api

import (
	"encoding/base64"
	mmerror "micro-mart/pkg/mm_error"
	"micro-mart/pkg/mmotel"
	"micro-mart/pkg/responder"
	"micro-mart/pkg/utils"
	"micro-mart/services/frontend_api/internal/model/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"micro-mart/services/frontend_api/internal/infrastructure/grpc_client"
	"micro-mart/services/frontend_api/internal/model/request"
)

type UserHandler struct {
	userClient *grpc_client.UserClient
	authClient *grpc_client.AuthClient
}

func NewUserHandler(userClient *grpc_client.UserClient, authClient *grpc_client.AuthClient) *UserHandler {
	return &UserHandler{
		userClient: userClient,
		authClient: authClient,
	}
}

// Register 處理用戶註冊請求
func (h *UserHandler) Register(c *gin.Context) {
	// 解析請求
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		mmErr := mmerror.New(mmerror.AccountPasswordError, "Invalid request:", err)
		mmotel.Warn(c.Request.Context(), "Invalid request: "+err.Error())
		responder.Error(mmErr).WithContext(c)
		return
	}

	ctx := c.Request.Context()

	userResp, err := h.userClient.Register(ctx, &req)

	if err != nil {
		responder.Error(err).WithContext(c)
		return
	}
	authResp, err := h.authClient.GenerateTokens(ctx, userResp.UserId, userResp.Username, userResp.Email, userResp.Role)

	if err != nil {
		responder.Error(err).WithContext(c)
		return
	}
	// 設定 HttpOnly + Secure Cookie
	refreshToken := authResp.RefreshToken
	// 正式環境 secure 才設定為 true (https)
	//c.SetCookie("refresh_token", refreshToken, 30*24*60*60, "/", "", true, true) // 30天, Secure + HttpOnly
	c.SetCookie("refresh_token", refreshToken, 30*24*60*60, "/", "", false, true) // 30天, Secure + HttpOnly

	// 回傳其餘資料（AccessToken 或其他資訊）
	responder.Ok(response.GetAccessTokenResponse{
		Success:     true,
		AccessToken: authResp.AccessToken,
	}).WithContext(c)
}

func (h *UserHandler) GenerateRandomKey(c *gin.Context) {
	// 解析請求
	var req request.GenerateRandomKeyRequest

	// 綁定查詢參數到結構體
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	// 依照request 的數字來創建key的大小
	jwtKey, err := utils.GenerateRandomKey(req.Length)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Generate random key failed: " + err.Error(),
		})
		return
	}

	// 用base64 輸出密鑰
	base64JwtKey := base64.StdEncoding.EncodeToString(jwtKey)
	c.JSON(http.StatusOK, base64JwtKey)
}
