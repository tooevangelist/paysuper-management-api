package model

const (
	DefaultLimit  = 100
	DefaultOffset = 0

	EmptyString = ""

	QueryParameterNameLimit  = "limit"
	QueryParameterNameOffset = "offset"
	QueryParameterNameSort   = "sort[]"

	ResponseMessageInvalidRequestData = "Invalid request data"
	ResponseMessageAccessDenied       = "Access denied"
	ResponseMessageNotFound           = "Not found"
	ResponseMessageProjectIdIsInvalid = "one or more project identifier is invalid"
	ResponseMessageUnknownDbError     = "err: 1, unknown error. try request later"
	ResponseMessageUnknownError       = "unknown error. try request later"
)

var DefaultSort = []string{"_id"}
