package core

import "fmt"

// ErrorCode is a stable machine-readable error identifier. / ErrorCode 是稳定的机器可读错误标识。
type ErrorCode string

const (
	ErrorCodeUnknown          ErrorCode = "ERR_UNKNOWN"
	ErrorCodeI18NLoad         ErrorCode = "ERR_I18N_LOAD"
	ErrorCodeRender           ErrorCode = "ERR_RENDER"
	ErrorCodeApplyPrecheck    ErrorCode = "ERR_APPLY_PRECHECK"
	ErrorCodeApplyVerify      ErrorCode = "ERR_APPLY_VERIFY"
	ErrorCodeStateRead        ErrorCode = "ERR_STATE_READ"
	ErrorCodeStateWrite       ErrorCode = "ERR_STATE_WRITE"
	ErrorCodeSchemaValidation ErrorCode = "ERR_SCHEMA_VALIDATION"
	ErrorCodeBackupNotFound   ErrorCode = "ERR_BACKUP_NOT_FOUND"
	ErrorCodeRollback         ErrorCode = "ERR_ROLLBACK"
	ErrorCodeLogRead          ErrorCode = "ERR_LOG_READ"
	ErrorCodeLogWrite         ErrorCode = "ERR_LOG_WRITE"
	ErrorCodeLogBundle        ErrorCode = "ERR_LOG_BUNDLE"
)

// AppError keeps stable codes separate from localized text. / AppError 将稳定错误码与本地化文本分离。
type AppError struct {
	Code       ErrorCode
	MessageKey string
	Cause      error
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}

	if e.Cause == nil {
		return fmt.Sprintf("%s (%s)", e.MessageKey, e.Code)
	}

	return fmt.Sprintf("%s (%s): %v", e.MessageKey, e.Code, e.Cause)
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Cause
}

// ExitError carries a process exit code without forcing duplicate stderr output. / ExitError 携带进程退出码且避免重复打印错误输出。
type ExitError struct {
	Code   int
	Silent bool
	Cause  error
}

func (e *ExitError) Error() string {
	if e == nil || e.Cause == nil {
		return ""
	}

	return e.Cause.Error()
}

func (e *ExitError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Cause
}
