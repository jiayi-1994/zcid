package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeSuccess             = 0
	CodeInternalServerError = 50001
)

// ContextKeyRequestID is the Gin context key for request ID.
const ContextKeyRequestID = "requestId"

type successBody struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data"`
	RequestID string `json:"requestId"`
}

type errorBody struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Detail    string `json:"detail,omitempty"`
	RequestID string `json:"requestId"`
}

func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, successBody{
		Code:      CodeSuccess,
		Message:   "success",
		Data:      data,
		RequestID: GetRequestID(c),
	})
}

func Error(c *gin.Context, httpStatus int, code int, message, detail string) {
	if httpStatus <= 0 {
		httpStatus = HTTPStatusFromCode(code)
	}

	c.JSON(httpStatus, errorBody{
		Code:      code,
		Message:   message,
		Detail:    detail,
		RequestID: GetRequestID(c),
	})
}

func HandleError(c *gin.Context, err error) {
	if err == nil {
		Success(c, nil)
		return
	}

	var bizErr *BizError
	if errors.As(err, &bizErr) {
		Error(c, bizErr.HTTPStatus, bizErr.Code, bizErr.Message, bizErr.Detail)
		return
	}

	Error(c, http.StatusInternalServerError, CodeInternalServerError, "internal server error", "")
}

func GetRequestID(c *gin.Context) string {
	if c == nil {
		return ""
	}
	requestID, exists := c.Get(ContextKeyRequestID)
	if !exists {
		return ""
	}
	value, ok := requestID.(string)
	if !ok {
		return ""
	}
	return value
}
