package response

import "net/http"

const (
	// Generic 400xx codes
	CodeBadRequest = 40001
	CodeValidation = 40002

	// Auth 401xx codes (Epic 2)
	CodeUnauthorized    = 40101
	CodeTokenExpired    = 40102
	CodeAccountDisabled = 40103

	// Permission 403xx codes (Epic 2)
	CodeForbidden = 40301

	// PipelineRun 4031x codes (Epic 7)
	CodeRunNotFound     = 40310
	CodeRunAlreadyDone  = 40311
	CodeRunConcurrency  = 40312
	CodeRunSubmitFailed = 40313
	CodeRunCancelFailed = 40314

	// Not found 404xx codes
	CodeNotFound = 40401

	// Conflict 409xx codes
	CodeConflict = 40901

	// Git integration 4041x codes (Epic 5)
	CodeGitTokenInvalid        = 40411
	CodeGitWebhookSigInvalid   = 40412
	CodeGitConnectionNotFound  = 40413
	CodeGitConnectionDown      = 40414
	CodeGitProviderUnsupported = 40415
	CodeGitAPIFailed           = 40416
	CodeGitNameDuplicate       = 40417

	// Pipeline 402xx codes (Epic 6)
	CodePipelineNotFound    = 40201
	CodePipelineNameDup     = 40202
	CodePipelineCRDFailed   = 40203
	CodePipelineConcurrency = 40204

	// Variable 405xx codes
	CodeVarDuplicate  = 40501
	CodeDecryptFailed = 40502
	CodeEncryptFailed = 40503

	// Registry 406xx codes (Epic 7)
	CodeRegistryNotFound   = 40601
	CodeRegistryNameDup    = 40602
	CodeRegistryConnFailed = 40603

	// WebSocket 407xx codes (Epic 8)
	CodeWSConnectionLimit = 40701
	CodeWSAuthFailed      = 40702
	CodeWSInvalidMessage  = 40703

	// Deployment 408xx codes (Epic 9)
	CodeDeployNotFound    = 40801
	CodeDeployFailed      = 40802
	CodeDeploySyncFailed  = 40803
	CodeDeployRollbackErr = 40804

	// Notification 409xx codes (Epic 10)
	CodeNotifRuleNotFound = 40902
	CodeNotifSendFailed   = 40903

	// Audit 410xx codes
	CodeAuditQueryFailed = 41001

	// Server 500xx codes
	CodeDependencyUnhealthy = 50002
)

var codeHTTPStatusMap = map[int]int{
	CodeBadRequest:             http.StatusBadRequest,
	CodeValidation:             http.StatusBadRequest,
	CodeUnauthorized:           http.StatusUnauthorized,
	CodeTokenExpired:           http.StatusUnauthorized,
	CodeAccountDisabled:        http.StatusUnauthorized,
	CodeForbidden:              http.StatusForbidden,
	CodeNotFound:               http.StatusNotFound,
	CodeConflict:               http.StatusConflict,
	CodeGitTokenInvalid:        http.StatusUnauthorized,
	CodeGitWebhookSigInvalid:   http.StatusUnauthorized,
	CodeGitConnectionNotFound:  http.StatusNotFound,
	CodeGitConnectionDown:      http.StatusBadRequest,
	CodeGitProviderUnsupported: http.StatusBadRequest,
	CodeGitAPIFailed:           http.StatusBadGateway,
	CodeGitNameDuplicate:       http.StatusConflict,
	CodePipelineNotFound:       http.StatusNotFound,
	CodePipelineNameDup:        http.StatusConflict,
	CodePipelineCRDFailed:      http.StatusInternalServerError,
	CodePipelineConcurrency:    http.StatusConflict,
	CodeRunNotFound:            http.StatusNotFound,
	CodeRunAlreadyDone:         http.StatusBadRequest,
	CodeRunConcurrency:         http.StatusConflict,
	CodeRunSubmitFailed:        http.StatusInternalServerError,
	CodeRunCancelFailed:        http.StatusInternalServerError,
	CodeVarDuplicate:           http.StatusConflict,
	CodeDecryptFailed:          http.StatusInternalServerError,
	CodeEncryptFailed:          http.StatusInternalServerError,
	CodeRegistryNotFound:       http.StatusNotFound,
	CodeRegistryNameDup:        http.StatusConflict,
	CodeRegistryConnFailed:     http.StatusBadGateway,
	CodeWSConnectionLimit:      http.StatusTooManyRequests,
	CodeWSAuthFailed:           http.StatusUnauthorized,
	CodeWSInvalidMessage:       http.StatusBadRequest,
	CodeDeployNotFound:         http.StatusNotFound,
	CodeDeployFailed:           http.StatusInternalServerError,
	CodeDeploySyncFailed:       http.StatusBadGateway,
	CodeDeployRollbackErr:      http.StatusBadRequest,
	CodeNotifRuleNotFound:      http.StatusNotFound,
	CodeNotifSendFailed:        http.StatusBadGateway,
	CodeAuditQueryFailed:       http.StatusInternalServerError,
	CodeInternalServerError:    http.StatusInternalServerError,
	CodeDependencyUnhealthy:    http.StatusServiceUnavailable,
}

func HTTPStatusFromCode(code int) int {
	if code == CodeSuccess {
		return http.StatusOK
	}
	if status, ok := codeHTTPStatusMap[code]; ok {
		return status
	}
	if code >= 40000 && code < 50000 {
		return http.StatusBadRequest
	}
	if code >= 50000 && code < 60000 {
		return http.StatusInternalServerError
	}
	return http.StatusInternalServerError
}
