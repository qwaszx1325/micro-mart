package responder

import (
	"github.com/gin-gonic/gin"
	mmerror "micro-mart/pkg/mm_error"
)

type Response struct {
	code    mmerror.MmCode
	message string
	data    any
}

// 200 回應
func Ok(data any) *Response {
	return &Response{
		code: mmerror.OK,
		data: data,
	}
}

// error 回應
func Error(err *mmerror.MmError) *Response {
	return &Response{
		code:    err.Code(),
		message: err.Message(),
		data:    err.Data(),
	}
}

// unknow error
func UnknownError(err error) *Response {
	return &Response{
		code:    mmerror.InternalServerError,
		message: err.Error(),
		data:    nil,
	}
}

// toGinH converts the Response to a Gin H map for JSON serialization.
//
// Parameters:
//   - traceId: A string representing the trace ID for the
//     current request.
//
// Returns:
//   - A Gin H map representing the Response object.
func (r *Response) toGinH(traceId string) gin.H {
	return gin.H{
		"code":    r.code,
		"message": r.message,
		"traceId": traceId,
		"data":    r.data,
	}
}

// HttpCode returns the HTTP status code associated with the Response.
//
// Returns:
//   - An integer representing the HTTP status code.
func (r *Response) HttpCode() int {
	return r.code.HttpCode()
}

// WithContext sets the Response in the given Gin context.
//
// This method allows storing the Response object in the Gin context
// for later retrieval and processing by middleware or other handlers.
//
// Parameters:
//   - c: A pointer to the gin.Context in which to store the Response.
//
// Usage:
//
//	res := responder.Ok(someData)
//	res.WithContext(c)
func (r *Response) WithContext(c *gin.Context) {
	c.Set(_responseKey, r)
}
