package mmerror

import (
	"encoding/json"
	"errors"
	"fmt"
	internal "micro-mart/pkg/mm_error/internal/gen"
	"net/http"

	"google.golang.org/grpc/status"
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

// Is checks if the target error matches the MmError.
func (e *MmError) Is(target error) bool {
	t, ok := target.(*MmError)
	if !ok {
		return false
	}
	return e.code == t.code
}

// Data returns the data associated with the MmError.
func (e *MmError) Data() any {
	return e.data
}

// WithSource add error sources to the MmError.
func (e *MmError) WithSource(err error) *MmError {
	e.sources = append(e.sources, err)
	return e
}

// FromGrpcErr converts a gRPC error to a MmError.
// Parameters:
//   - err: The gRPC error.
//
// Returns:
//   - error: The MmError.
//   - ok: A boolean indicating if the conversion was successful.
//
// Example:
//
//	mmErr, ok := FromGrpcErr(err)
func FromGrpcErr(err error) (mmErr *MmError, ok bool) {
	st, ok := status.FromError(err)
	if !ok {
		return nil, false
	}

	// Check if the error is our custom MmError
	for _, detail := range st.Details() {
		if proto, ok := detail.(*internal.MmErrorProto); ok {
			mmErr, err := fromProto(proto)
			if err != nil {
				return nil, false
			}
			return mmErr, true
		}
	}

	return nil, false
}

// toProto converts the MmError to a proto message.
func (e *MmError) toProto() (*internal.MmErrorProto, error) {
	dataBytes, err := json.Marshal(e.data)
	if err != nil {
		return nil, err
	}

	sources := make([]string, len(e.sources))
	for i, src := range e.sources {
		if src != nil {
			sources[i] = src.Error()
		}
	}

	return &internal.MmErrorProto{
		Code:    int32(e.code),
		Message: e.msg,
		Data:    dataBytes,
		Source:  sources,
	}, nil
}

// fromProto converts a proto message to a MmError.
func fromProto(proto *internal.MmErrorProto) (*MmError, error) {
	data := make(map[string]interface{})
	if err := json.Unmarshal(proto.Data, &data); err != nil {
		return nil, err
	}

	sources := make([]error, len(proto.Source))
	for i, src := range proto.Source {
		sources[i] = errors.New(src)
	}

	return &MmError{
		code:    MmCode(proto.Code),
		msg:     proto.Message,
		data:    data,
		sources: sources,
	}, nil
}
