package apperrors

const (
	CodeInternal = "INTERNAL_ERROR"

	CodeReadyDatabasePingFailed = "READY_DATABASE_PING_FAILED"
	CodeReadySQLGetterFailed    = "READY_SQL_GETTER_FAILED"

	CodeInternalUnauthorized = "INTERNAL_UNAUTHORIZED"
)

const (
	MsgInternal = "An unexpected error occurred."

	MsgReadyDatabasePingFailed = "Database is not reachable."
	MsgReadySQLGetterFailed    = "Database connection failed."

	MsgInternalUnauthorized = "Invalid internal token."
)
