package response

import (
	"fmt"
)

// BizError represents business-level errors with stable code mapping.
type BizError struct {
	Code       int
	Message    string
	Detail     string
	HTTPStatus int
}

func (e *BizError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Detail)
	}
	return e.Message
}

func NewBizError(code int, message, detail string) *BizError {
	return &BizError{
		Code:       code,
		Message:    message,
		Detail:     detail,
		HTTPStatus: HTTPStatusFromCode(code),
	}
}

func NewBizErrorWithStatus(code int, httpStatus int, message, detail string) *BizError {
	status := httpStatus
	if status <= 0 {
		status = HTTPStatusFromCode(code)
	}
	return &BizError{
		Code:       code,
		Message:    message,
		Detail:     detail,
		HTTPStatus: status,
	}
}
