package qac

import (
	"errors"
	"fmt"
	"testing"
)

func TestAddErrorfIsAssertionFailure(t *testing.T) {
	var r AssertionResult
	r.addErrorf("expected %d got %d", 1, 2)
	var qe *QacError
	if !errors.As(r.errors[0], &qe) {
		t.Fatal("expected *QacError")
	}
	if qe.Kind != KindAssertionFailure {
		t.Errorf("expected KindAssertionFailure, got %v", qe.Kind)
	}
}

func TestAddInfraErrorIsInfrastructure(t *testing.T) {
	var r AssertionResult
	r.addInfraError(fmt.Errorf("read /tmp/x: permission denied"))
	var qe *QacError
	if !errors.As(r.errors[0], &qe) {
		t.Fatal("expected *QacError")
	}
	if qe.Kind != KindInfrastructure {
		t.Errorf("expected KindInfrastructure, got %v", qe.Kind)
	}
}

func TestAddConfigErrorIsConfiguration(t *testing.T) {
	var r AssertionResult
	r.addConfigError(fmt.Errorf("strconv.Atoi: parsing \"abc\""))
	var qe *QacError
	if !errors.As(r.errors[0], &qe) {
		t.Fatal("expected *QacError")
	}
	if qe.Kind != KindConfiguration {
		t.Errorf("expected KindConfiguration, got %v", qe.Kind)
	}
}

func TestAddErrorWrapsPlainAsAssertionFailure(t *testing.T) {
	var r AssertionResult
	r.addError(fmt.Errorf("plain error"))
	var qe *QacError
	if !errors.As(r.errors[0], &qe) {
		t.Fatal("expected plain error to be wrapped as *QacError")
	}
	if qe.Kind != KindAssertionFailure {
		t.Errorf("expected KindAssertionFailure for plain error, got %v", qe.Kind)
	}
}

func TestAddErrorPreservesExistingQacErrorKind(t *testing.T) {
	var r AssertionResult
	original := asInfraError(fmt.Errorf("disk full"))
	r.addError(original)
	var qe *QacError
	if !errors.As(r.errors[0], &qe) {
		t.Fatal("expected *QacError")
	}
	if qe.Kind != KindInfrastructure {
		t.Errorf("expected kind to be preserved as KindInfrastructure, got %v", qe.Kind)
	}
}

func TestAddErrorsWrapsEachElement(t *testing.T) {
	var r AssertionResult
	r.addErrors([]error{fmt.Errorf("err1"), fmt.Errorf("err2")})
	if len(r.errors) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(r.errors))
	}
	for i, err := range r.errors {
		var qe *QacError
		if !errors.As(err, &qe) {
			t.Errorf("error %d: expected *QacError", i)
		}
	}
}

func TestQacErrorUnwrapChain(t *testing.T) {
	cause := fmt.Errorf("original cause")
	wrapped := fmt.Errorf("context: %w", cause)
	qe := asInfraError(wrapped)
	if !errors.Is(qe, cause) {
		t.Error("errors.Is should find cause through QacError.Unwrap chain")
	}
}

func TestAsConfigErrorPreservesMessage(t *testing.T) {
	inner := fmt.Errorf("yaml: line 5: cannot unmarshal")
	qe := asConfigError(inner)
	if qe.Error() != inner.Error() {
		t.Errorf("expected message %q, got %q", inner.Error(), qe.Error())
	}
	if qe.Kind != KindConfiguration {
		t.Errorf("expected KindConfiguration, got %v", qe.Kind)
	}
}
