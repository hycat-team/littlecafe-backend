package i18n

// Message keys are the stable identifiers shared by message bundles and call
// sites. Always reference these constants instead of hardcoding key strings.
const (
	// Error titles by HTTP status.
	ErrorTitleBadRequest            = "error.title.bad_request"
	ErrorTitleUnauthorized          = "error.title.unauthorized"
	ErrorTitleForbidden             = "error.title.forbidden"
	ErrorTitleNotFound              = "error.title.not_found"
	ErrorTitleConflict              = "error.title.conflict"
	ErrorTitleTooManyRequests       = "error.title.too_many_requests"
	ErrorTitleRequestEntityTooLarge = "error.title.request_entity_too_large"
	ErrorTitleInternal              = "error.title.internal"
	ErrorTitleServiceUnavailable    = "error.title.service_unavailable"

	// Generic client-safe messages.
	ErrorInternalGeneric = "error.internal.generic"
	ErrorInternalPanic   = "error.internal.panic"

	// Authentication and authorization.
	ErrorAuthInvalidToken   = "error.auth.invalid_token"
	ErrorAuthSessionExpired = "error.auth.session_expired"
	ErrorAuthRequired       = "error.auth.required"
	ErrorForbiddenAccess    = "error.forbidden.access"

	// Infrastructure and upstream.
	ErrorSearchIndexNotFound = "error.search_index.not_found"
	ErrorUpstreamUnavailable = "error.upstream.unavailable"
	ErrorRateLimitExceeded   = "error.rate_limit.exceeded"
	ErrorLoginRateLimit      = "error.login.rate_limit_exceeded"
	ErrorInvalidUUID         = "error.invalid_uuid"
	ErrorRealtimeUnsupported = "error.realtime.unsupported"
	ErrorJWTMissingPublicKey = "error.jwt.missing_public_key"

	// Validation envelope.
	ErrorValidationInvalidFormatTitle  = "error.validation.invalid_format.title"
	ErrorValidationInvalidFormatDetail = "error.validation.invalid_format.detail"

	// Validation rules.
	ValidationRequired           = "validation.required"
	ValidationEmail              = "validation.email"
	ValidationUsername           = "validation.username"
	ValidationPasswordComplexity = "validation.password_complexity"
	ValidationEqField            = "validation.eqfield"
	ValidationNeField            = "validation.nefield"
	ValidationLen                = "validation.len"
	ValidationGT                 = "validation.gt"
	ValidationMin                = "validation.min"
	ValidationMaxInt             = "validation.max_int"
	ValidationMaxStr             = "validation.max_str"
	ValidationDateTime           = "validation.datetime"
	ValidationOneOf              = "validation.oneof"
	ValidationNumeric            = "validation.numeric"
	ValidationDefault            = "validation.default"
	ValidationBodyTooLarge       = "validation.body_too_large"
	ValidationParamMissing       = "validation.param.missing"
	ValidationParamInvalid       = "validation.param.invalid"

	// Success messages.
	SuccessOK = "success.ok"
)

// FieldLabelKey returns the catalog key for a client-facing field label.
func FieldLabelKey(jsonName string) string {
	return "field." + jsonName
}
