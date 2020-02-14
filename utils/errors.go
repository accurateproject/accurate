package utils

import (
	"errors"
	"fmt"
)

var (
	ErrNotImplemented          = errors.New("NOT_IMPLEMENTED")
	ErrNotFound                = errors.New("NOT_FOUND")
	ErrTimedOut                = errors.New("TIMED_OUT")
	ErrServerError             = errors.New("SERVER_ERROR")
	ErrQuotaExceeded           = errors.New("QUOTA_EXCEEDED")
	ErrMaxRecursionDepth       = errors.New("MAX_RECURSION_DEPTH")
	ErrMandatoryIeMissing      = errors.New("MANDATORY_IE_MISSING")
	ErrExists                  = errors.New("EXISTS")
	ErrBrokenReference         = errors.New("BROKEN_REFERENCE")
	ErrParserError             = errors.New("PARSER_ERROR")
	ErrInvalidPath             = errors.New("INVALID_PATH")
	ErrInvalidKey              = errors.New("INVALID_KEY")
	ErrUnauthorizedDestination = errors.New("UNAUTHORIZED_DESTINATION")
	ErrRatingPlanNotFound      = errors.New("RATING_PLAN_NOT_FOUND")
	ErrAccountNotFound         = errors.New("ACCOUNT_NOT_FOUND")
	ErrAccountDisabled         = errors.New("ACCOUNT_DISABLED")
	ErrUserNotFound            = errors.New("USER_NOT_FOUND")
	ErrInsufficientCredit      = errors.New("INSUFFICIENT_CREDIT")
	ErrNotConvertible          = errors.New("NOT_CONVERTIBLE")
	ErrResourceUnavailable     = errors.New("RESOURCE_UNAVAILABLE")
	ErrNoActiveSession         = errors.New("NO_ACTIVE_SESSION")
)

// NewCGRError initialises a new CGRError
func NewCGRError(context, apiErr, shortErr, longErr string) *CGRError {
	return &CGRError{context: context, apiError: apiErr,
		shortError: shortErr, longError: longErr, errorMessage: shortErr}
}

// CGRError is a context based error
// returns always errorMessage but this can be switched based on methods called on it
type CGRError struct {
	context      string
	apiError     string
	shortError   string
	longError    string
	errorMessage string
}

func (err *CGRError) Context() string {
	return err.context
}

func (err *CGRError) Error() string {
	return err.errorMessage
}

func (err *CGRError) ActivateAPIError() {
	err.errorMessage = err.apiError
}

func (err *CGRError) ActivateShortError() {
	err.errorMessage = err.shortError
}

func (err *CGRError) ActivateLongError() {
	err.errorMessage = err.longError
}

func NewErrMandatoryIeMissing(fields ...string) error {
	return fmt.Errorf("MANDATORY_IE_MISSING:%v", fields)
}

func NewErrServerError(err error) error {
	return fmt.Errorf("SERVER_ERROR: %s", err)
}

// Centralized returns for APIs
func APIErrorHandler(err error) error {
	cgrErr, ok := err.(*CGRError)
	if !ok {
		if err == ErrNotFound {
			return err
		} else {
			return NewErrServerError(err)
		}
	}
	cgrErr.ActivateAPIError()
	return cgrErr
}
