package mmerror

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

type MmCode int

const (
	// 200 status code from here
	OK = 200_0000 // 成功
	//InvalidArgument

	// 400 status code from here
	AccountError            = 400_0000 // 帳號錯誤
	AccountPasswordError    = 400_0001 // 密碼錯誤
	InvalidVerificationCode = 400_0002 // 驗證碼錯誤
	InvalidArgument         = 400_0003
	BadRequest              = 400_0098 // 請求錯誤
	ThirdPartyError         = 400_0099 // 第三方錯誤

	// 401 status code from here
	Unauthorized       = 401_0000 // 未授權
	UnusualLogin       = 401_0001 // 登入異常
	AccountLocked      = 401_0002 // 帳號被鎖定
	WrongPassword      = 401_0003 // 密碼錯誤
	TokenExpired       = 401_0004 // Token過期
	ClientInactive     = 401_0005 // 客戶端未啟用
	MissingAccessToken = 401_0006
	// 403 status code from here
	Forbidden    = 403_0000 // 禁止訪問
	NoPermission = 403_0001 // 沒有權限
	NoRole       = 403_0002 // 沒有角色

	// 404 status code from here
	ResponseNotFound = 404_0000 // 沒有Response
	ResourceNotFound = 404_0001 // 找不到資源

	// 409 status code from here
	Conflict        = 409_0000 // 衝突
	ResourceIsExist = 409_0001 // 資源已存在

	// 429 status code from here
	TooManyRequests = 429_0000 // 請求過多

	// 500 status code from here
	InternalServerError = 500_0000 // 内部錯誤
	InvalidPermission   = 500_0001 // 無效的權限

	// 501 status code from here
	NotImplemented = 501_0000 // 功能未實現
)

// HttpCode returns the standard HTTP status code.
func (c MmCode) HttpCode() int {

	// Get the http code from the MmCode
	httpCode := int(c) / 10000

	// Check if the http code is valid
	if http.StatusText(httpCode) == "" {
		return http.StatusInternalServerError
	}

	return httpCode
}

func (c MmCode) GrpcCode() codes.Code {
	switch c.HttpCode() {
	case http.StatusOK:
		return codes.OK
	case http.StatusBadRequest:
		return codes.InvalidArgument
	case http.StatusUnauthorized:
		return codes.Unauthenticated
	case http.StatusForbidden:
		return codes.PermissionDenied
	case http.StatusNotFound:
		return codes.NotFound
	case http.StatusConflict:
		return codes.AlreadyExists
	case http.StatusTooManyRequests:
		return codes.ResourceExhausted
	case http.StatusInternalServerError:
		return codes.Internal
	case http.StatusNotImplemented:
		return codes.Unimplemented
	default:
		return codes.Unknown
	}
}

// Int returns the integer value of the MmCode.
func (c MmCode) Int() int {
	return int(c)
}
