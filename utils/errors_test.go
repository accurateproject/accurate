package utils

import (
	"testing"
)

func TestCGRErrorActivate(t *testing.T) {
	ctx := "TEST_CONTEXT"
	apiErr := "TEST_API_ERR"
	shortErr := "short error"
	longErr := "long error which is good for debug"
	err := NewCGRError(ctx, apiErr, shortErr, longErr)
	if ctxRcv := err.Context(); ctxRcv != ctx {
		t.Errorf("Context: <%s>", ctxRcv)
	}
	if err.Error() != shortErr {
		t.Error(err)
	}
	err.ActivateAPIError()
	if err.Error() != apiErr {
		t.Error(err)
	}
	err.ActivateLongError()
	if err.Error() != longErr {
		t.Error(err)
	}
}

func TestAPIErrorHandler(t *testing.T) {
	if err := APIErrorHandler(ErrNotImplemented); err.Error() != NewErrServerError(ErrNotImplemented).Error() {
		t.Error(err)
	}
	if err := APIErrorHandler(ErrNotFound); err.Error() != ErrNotFound.Error() {
		t.Error(err)
	}
	cgrErr := NewCGRError("TEST_CONTEXT", "TEST_API_ERR", "short error", "long error which is good for debug")
	if err := APIErrorHandler(cgrErr); err.Error() != cgrErr.apiError {
		t.Error(err)
	}
}
