package error

import (
	"fmt"
	"net/http"
)

// MmError is a custom error type that wraps error code, message, and error sources.
type MmError struct {
	code    MmCode
	msg     string
	data    any
	sources []error
}

// New creates a new MmError.
// Parameters:
//   - code: The MmCode of the error.
//   - msg: The error message.
//   - data: The data associated with the error.
//   - source: The error sources.
//
// Returns:
//   - error: The MmError.
//
// Example:
//
//	err := NewMmError(ErrInvalidInput, "err_msg")
//	err := NewMmError(ErrInvalidInput, "err_msg",err)
//	err := NewMmError(ErrInvalidInput, "err_msg",err1,err2)
func New(code MmCode, msg string, source ...error) *MmError {
	return &MmError{
		code:    code,
		msg:     msg,
		sources: source,
	}
}

func (e *MmError) Error() string {
	if len(e.sources) > 0 {
		return fmt.Sprintf("mmCode: %v, msg:%s,  sources: %v", e.code, e.msg, e.sources)
	}
	return fmt.Sprintf("mmCode: %v, msg:%s", e.code, e.msg)
}

// HttpCode returns the standard HTTP status code.
func (e *MmError) HttpCode() int {
	// Check self is nil
	if e == nil {
		return http.StatusInternalServerError
	}

	return e.code.HttpCode()
}

// Code returns the mm error code.
func (e *MmError) Code() MmCode {
	return e.code
}

// Message returns the error message.
func (e *MmError) Message() string {
	return e.msg
}

// Unwrap returns the error sources.
// if there are no sources, it will return nil.
func (e *MmError) Unwrap() []error {
	return e.sources
}

// WithData add data to the MmError.
// Parameters:
//   - data: The data to be added to the MmError.
//
// Returns:
//   - error: The MmError.
//
// Example:
//
//	data := map[string]interface{}{"key1": "value1"}
//	err := NewMmError(ErrInvalidInput, "err_msg").WithData(data)
func (e *MmError) WithData(data any) *MmError {
	e.data = data
	return e
}
