package httperr

import "net/http"

// stdErr represents a minimal error for an HTTP status code.
type stdErr int

// Common HTTP errors.
const (
	ErrBadRequest                   stdErr = http.StatusBadRequest
	ErrUnauthorized                 stdErr = http.StatusUnauthorized
	ErrPaymentRequired              stdErr = http.StatusPaymentRequired
	ErrForbidden                    stdErr = http.StatusForbidden
	ErrNotFound                     stdErr = http.StatusNotFound
	ErrMethodNotAllowed             stdErr = http.StatusMethodNotAllowed
	ErrNotAcceptable                stdErr = http.StatusNotAcceptable
	ErrProxyAuthRequired            stdErr = http.StatusProxyAuthRequired
	ErrRequestTimeout               stdErr = http.StatusRequestTimeout
	ErrConflict                     stdErr = http.StatusConflict
	ErrGone                         stdErr = http.StatusGone
	ErrLengthRequired               stdErr = http.StatusLengthRequired
	ErrPreconditionFailed           stdErr = http.StatusPreconditionFailed
	ErrRequestEntityTooLarge        stdErr = http.StatusRequestEntityTooLarge
	ErrRequestURITooLong            stdErr = http.StatusRequestURITooLong
	ErrUnsupportedMediaType         stdErr = http.StatusUnsupportedMediaType
	ErrRequestedRangeNotSatisfiable stdErr = http.StatusRequestedRangeNotSatisfiable
	ErrExpectationFailed            stdErr = http.StatusExpectationFailed
	ErrTeapot                       stdErr = http.StatusTeapot
	ErrUnprocessableEntity          stdErr = http.StatusUnprocessableEntity
	ErrLocked                       stdErr = http.StatusLocked
	ErrFailedDependency             stdErr = http.StatusFailedDependency
	ErrUpgradeRequired              stdErr = http.StatusUpgradeRequired
	ErrPreconditionRequired         stdErr = http.StatusPreconditionRequired
	ErrTooManyRequests              stdErr = http.StatusTooManyRequests
	ErrRequestHeaderFieldsTooLarge  stdErr = http.StatusRequestHeaderFieldsTooLarge
	ErrUnavailableForLegalReasons   stdErr = http.StatusUnavailableForLegalReasons

	ErrInternalServerError                 stdErr = http.StatusInternalServerError
	ErrNotImplemented                      stdErr = http.StatusNotImplemented
	ErrBadGateway                          stdErr = http.StatusBadGateway
	ErrServiceUnavailable                  stdErr = http.StatusServiceUnavailable
	ErrGatewayTimeout                      stdErr = http.StatusGatewayTimeout
	ErrStatusHTTPVersionNotSupported       stdErr = http.StatusHTTPVersionNotSupported
	ErrStatusVariantAlsoNegotiates         stdErr = http.StatusVariantAlsoNegotiates
	ErrStatusInsufficientStorage           stdErr = http.StatusInsufficientStorage
	ErrStatusLoopDetected                  stdErr = http.StatusLoopDetected
	ErrStatusNotExtended                   stdErr = http.StatusNotExtended
	ErrStatusNetworkAuthenticationRequired stdErr = http.StatusNetworkAuthenticationRequired
)

var _ error = stdErr(0)
var _ statusCoder = stdErr(0)

func (e stdErr) Error() string {
	text := http.StatusText(int(e))
	if text == "" {
		return "Unknown Status"
	}
	return text
}

func (e stdErr) StatusCode() int {
	return int(e)
}
