package apperrors

type Kind int

const (
	KindInvalid Kind = iota
	KindNotFound
	KindForbidden
	KindUnauthorized
	KindConflict
	KindInternal
	KindUnavailable
)
