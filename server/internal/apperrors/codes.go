package apperrors

const (
	CodeInternal = "INTERNAL_ERROR"

	CodeReadyDatabasePingFailed = "READY_DATABASE_PING_FAILED"
	CodeReadySQLGetterFailed    = "READY_SQL_GETTER_FAILED"

	CodeInternalUnauthorized = "INTERNAL_UNAUTHORIZED"

	CodeAuthInvalidBody       = "AUTH_INVALID_BODY"
	CodeAuthEmailInvalid      = "AUTH_EMAIL_INVALID"
	CodeAuthPasswordWeak      = "AUTH_PASSWORD_WEAK"
	CodeAuthEmailTaken        = "AUTH_EMAIL_TAKEN"
	CodeAuthInvalidCredentials = "AUTH_INVALID_CREDENTIALS"
	CodeAuthUnauthorized      = "AUTH_UNAUTHORIZED"
	CodeAuthTokenInvalid      = "AUTH_TOKEN_INVALID"

	CodeProfileNotFound     = "PROFILE_NOT_FOUND"
	CodeProfileSlugInvalid  = "PROFILE_SLUG_INVALID"
	CodeProfileSlugTaken    = "PROFILE_SLUG_TAKEN"
	CodeProfileInvalidBody  = "PROFILE_INVALID_BODY"

	CodeExperienceNotFound = "EXPERIENCE_NOT_FOUND"
	CodeEducationNotFound  = "EDUCATION_NOT_FOUND"
	CodeInvalidID          = "INVALID_ID"
)

const (
	MsgInternal = "An unexpected error occurred."

	MsgReadyDatabasePingFailed = "Database is not reachable."
	MsgReadySQLGetterFailed    = "Database connection failed."

	MsgInternalUnauthorized = "Invalid internal token."

	MsgAuthEmailInvalid       = "Invalid email address."
	MsgAuthPasswordWeak       = "Password must be at least 8 characters."
	MsgAuthEmailTaken         = "Email is already registered."
	MsgAuthInvalidCredentials = "Invalid email or password."
	MsgAuthUnauthorized       = "Authentication required."
	MsgAuthTokenInvalid       = "Invalid or expired token."

	MsgProfileNotFound = "Profile not found."
	MsgProfileSlugTaken = "Slug is already taken."

	MsgExperienceNotFound = "Experience not found."
	MsgEducationNotFound  = "Education not found."
	MsgInvalidID          = "Invalid id."
)
