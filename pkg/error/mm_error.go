package error

import "fmt"

type Error struct {
	Code    string
	Message string
	Cause   error
}

// 錯誤碼常數
const (
	ResourceNotFound    = "RESOURCE_NOT_FOUND"
	InternalServerError = "INTERNAL_SERVER_ERROR"
)

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (原因: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}
