package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"micro-mart/pkg/mmotel"
	"micro-mart/services/frontend_api/internal/infrastructure/grpc_client"
	"micro-mart/services/frontend_api/internal/model/request"
)

type UserHandler struct {
	userClient *grpc_client.UserClient
}

func NewUserHandler(userClient *grpc_client.UserClient) *UserHandler {
	return &UserHandler{
		userClient: userClient,
	}
}

// Register 處理用戶註冊請求
func (h *UserHandler) Register(c *gin.Context) {

	// 解析請求
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	// 使用 mmotel 包裝 span
	ctx := c.Request.Context()
	var response interface{}
	var err error

	err = mmotel.WithSpan(ctx, "UserHandler.Register", func(ctx context.Context) error {
		// 調用 user client 的 Register 方法
		response, err = h.userClient.Register(ctx, &req)
		return err
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Registration failed: " + err.Error(),
		})
		return
	}

	// 返回註冊結果
	c.JSON(http.StatusOK, response)
}
