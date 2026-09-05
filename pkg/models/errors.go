package models

import "fmt"

type ErrorType string

const (
	ErrTypePermission   ErrorType = "PERMISSION_DENIED"
	ErrTypeNotFound     ErrorType = "NOT_FOUND"
	ErrTypeInvalidPath  ErrorType = "INVALID_PATH"
	ErrTypeIO           ErrorType = "IO_ERROR"
	ErrTypeCacheCorrupt ErrorType = "CACHE_CORRUPT"
)

type ApplicationError struct {
	Type      ErrorType
	Msg       string
	Path      string
	ErrOrigin error
}

func (e *ApplicationError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("%s: %s (path: %s)", e.Type, e.Msg, e.Path)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Msg)
}

func (e *ApplicationError) Unwrap() error {
	return e.ErrOrigin
}

func NewPermissionError(path string, err error) *ApplicationError {
	return &ApplicationError{
		Type:      ErrTypePermission,
		Msg:       "permission denied",
		Path:      path,
		ErrOrigin: err,
	}
}

func NewNotFoundError(path string, err error) *ApplicationError {
	return &ApplicationError{
		Type:      ErrTypeNotFound,
		Msg:       "file or directory not found",
		Path:      path,
		ErrOrigin: err,
	}
}
