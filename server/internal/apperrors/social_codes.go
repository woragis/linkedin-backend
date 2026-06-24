package apperrors

const (
	CodeConnectionNotFound      = "CONNECTION_NOT_FOUND"
	CodeConnectionInvalid       = "CONNECTION_INVALID"
	CodeConnectionExists        = "CONNECTION_EXISTS"
	CodeConnectionSelf          = "CONNECTION_SELF"
	CodeConnectionForbidden     = "CONNECTION_FORBIDDEN"

	CodePostNotFound    = "POST_NOT_FOUND"
	CodePostInvalidBody = "POST_INVALID_BODY"
	CodeCommentInvalid  = "COMMENT_INVALID_BODY"
	CodeCommentNotFound = "COMMENT_NOT_FOUND"
	CodeEventsInvalid   = "EVENTS_INVALID_BODY"
)

const (
	MsgConnectionNotFound  = "Connection not found."
	MsgConnectionExists    = "Connection already exists."
	MsgConnectionSelf      = "Cannot connect to yourself."
	MsgConnectionForbidden = "Not allowed on this connection."

	MsgPostNotFound = "Post not found."
	MsgCommentNotFound = "Comment not found."
)
