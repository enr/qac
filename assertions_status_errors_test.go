package qac

import (
	"errors"
	"testing"
)

// --- LesserThan missing failure cases ---

var statusLesserThanFailSpecs = []statusAssertionTestCase{
	{exitCode: 1, value: "1", expectedSuccess: false}, // equal: not strictly less
	{exitCode: 2, value: "1", expectedSuccess: false}, // greater: not less
}

func TestLesserThanAssertionFailure(t *testing.T) {
	for _, spec := range statusLesserThanFailSpecs {
		context := planContext{
			commandResult: executionResult{exitCode: spec.exitCode},
		}
		sut := &StatusAssertion{LesserThan: spec.value}
		r := sut.verify(context)
		if r.Success() != spec.expectedSuccess {
			t.Errorf("LesserThan %s: exit %d: expected success=%t got success=%t",
				spec.value, spec.exitCode, spec.expectedSuccess, r.Success())
		}
	}
}

// --- Invalid (non-numeric) field values ---

func TestStatusAssertionInvalidEqualsTo(t *testing.T) {
	ctx := planContext{commandResult: executionResult{exitCode: 0}}
	sut := &StatusAssertion{EqualsTo: "zero"}
	r := sut.verify(ctx)
	if r.Success() {
		t.Error("expected failure for non-numeric equals_to value")
	}
}

func TestStatusAssertionInvalidGreaterThan(t *testing.T) {
	ctx := planContext{commandResult: executionResult{exitCode: 5}}
	sut := &StatusAssertion{GreaterThan: "not-a-number"}
	r := sut.verify(ctx)
	if r.Success() {
		t.Error("expected failure for non-numeric greater_than value")
	}
}

func TestStatusAssertionInvalidLesserThan(t *testing.T) {
	ctx := planContext{commandResult: executionResult{exitCode: 0}}
	sut := &StatusAssertion{LesserThan: "not-a-number"}
	r := sut.verify(ctx)
	if r.Success() {
		t.Error("expected failure for non-numeric lesser_than value")
	}
}

// --- Command error propagation ---

func TestStatusAssertion_CommandError_NoConstraint(t *testing.T) {
	// When the command errored and no exit-code constraint is set,
	// the error must surface in the assertion result.
	ctx := planContext{
		commandResult: executionResult{
			exitCode: 1,
			err:      errors.New("process exited with status 1"),
		},
	}
	sut := &StatusAssertion{}
	r := sut.verify(ctx)
	if r.Success() {
		t.Error("expected failure when command errored with no exit-code constraints")
	}
}

func TestStatusAssertion_CommandError_EqualsToZero_NotAcceptable(t *testing.T) {
	// equals_to: "0" and exit code is 0 — the comparison passes,
	// but commandErrorIsAcceptable stays false so the error must still surface.
	ctx := planContext{
		commandResult: executionResult{
			exitCode: 0,
			err:      errors.New("some low-level error"),
		},
	}
	sut := &StatusAssertion{EqualsTo: "0"}
	r := sut.verify(ctx)
	if r.Success() {
		t.Error("expected failure: equals_to 0 does not make command errors acceptable")
	}
}

func TestStatusAssertion_CommandError_EqualsToNonZero_Acceptable(t *testing.T) {
	// equals_to: "1" means a non-zero exit is expected; command error is acceptable.
	ctx := planContext{
		commandResult: executionResult{
			exitCode: 1,
			err:      errors.New("non-zero exit"),
		},
	}
	sut := &StatusAssertion{EqualsTo: "1"}
	r := sut.verify(ctx)
	if !r.Success() {
		t.Errorf("expected success when exit matches expected non-zero value, got: %v", r.Errors())
	}
}

func TestStatusAssertion_CommandError_GreaterThan_Acceptable(t *testing.T) {
	ctx := planContext{
		commandResult: executionResult{
			exitCode: 2,
			err:      errors.New("non-zero exit"),
		},
	}
	sut := &StatusAssertion{GreaterThan: "1"}
	r := sut.verify(ctx)
	if !r.Success() {
		t.Errorf("expected success: greater_than makes command errors acceptable, got: %v", r.Errors())
	}
}
