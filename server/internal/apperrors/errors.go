package apperrors

import "errors"

type Error struct {
	Code    string
	Message string
	Kind    Kind
	Cause   error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Cause }

func Invalid(code, msg string) *Error {
	return &Error{Code: code, Message: msg, Kind: KindInvalid}
}

func NotFound(code, msg string) *Error {
	return &Error{Code: code, Message: msg, Kind: KindNotFound}
}

func Forbidden(code, msg string) *Error {
	return &Error{Code: code, Message: msg, Kind: KindForbidden}
}

func Unauthorized(code, msg string) *Error {
	return &Error{Code: code, Message: msg, Kind: KindUnauthorized}
}

func Conflict(code, msg string) *Error {
	return &Error{Code: code, Message: msg, Kind: KindConflict}
}

func InternalCause(code, msg string, cause error) *Error {
	return &Error{Code: code, Message: msg, Kind: KindInternal, Cause: cause}
}

func UnavailableCause(code, msg string, cause error) *Error {
	return &Error{Code: code, Message: msg, Kind: KindUnavailable, Cause: cause}
}

func As(err error) (*Error, bool) {
	var ae *Error
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}

func IsNotFound(err error) bool {
	ae, ok := As(err)
	return ok && ae.Kind == KindNotFound
}
