package mmerror

import (
	"context"
	"micro-mart/pkg/mmotel"
)

// LogAndReturnError 創建錯誤並記錄日誌，然後返回錯誤
// 參數：
//   - ctx: 上下文
//   - code: 錯誤代碼
//   - message: 錯誤消息
//   - err: 原始錯誤
//   - logMessage: 日誌消息
//   - fields: 附加字段
// 返回：
//   - *MmError: 創建的錯誤
func LogAndReturnError(ctx context.Context, code MmCode, message string, err error, logMessage string, fields ...mmotel.Field) *MmError {
	// 創建錯誤
	mmErr := New(code, message, err)

	// 創建附加字段
	extraFields := []mmotel.Field{
		mmotel.NewField("error", mmErr.ErrorWithStack()),
	}

	// 合併原有字段和附加字段
	allFields := append(extraFields, fields...)

	// 記錄日誌
	// 注意：這裡不需要 allFields...，因為 Error 已經接受可變參數
	mmotel.Error(ctx, logMessage, allFields...)

	return mmErr
}
