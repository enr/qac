package qac

import (
	"errors"
	"testing"
)

// --- LessThan missing failure cases ---

var statusLessThanFailSpecs = []statusAssertionTestCase{
	{exitCode: 1, value: 1, expectedSuccess: false}, // equal: not strictly less
	{exitCode: 2, value: 1, expectedSuccess: false}, // greater: not less
}

func TestLessThanAssertionFailure(t *testing.T) {
	for _, spec := range statusLessThanFailSpecs {
		context := planContext{
			commandResult: executionResult{exitCode: spec.exitCode},
		}
		sut := &StatusAssertion{LessThan: &spec.value}
		r := sut.verify(context)
		if r.Success() != spec.expectedSuccess {
			t.Errorf("LessThan %d: exit %d: expected success=%t got success=%t",
				spec.value, spec.exitCode, spec.expectedSuccess, r.Success())
		}
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
	// equals_to: 0 and exit code is 0 — the comparison passes,
	// but commandErrorIsAcceptable stays false so the error must still surface.
	zero := 0
	ctx := planContext{
		commandResult: executionResult{
			exitCode: 0,
			err:      errors.New("some low-level error"),
		},
	}
	sut := &StatusAssertion{EqualsTo: &zero}
	r := sut.verify(ctx)
	if r.Success() {
		t.Error("expected failure: equals_to 0 does not make command errors acceptable")
	}
}

func TestStatusAssertion_CommandError_EqualsToNonZero_Acceptable(t *testing.T) {
	// equals_to: 1 means a non-zero exit is expected; command error is acceptable.
	one := 1
	ctx := planContext{
		commandResult: executionResult{
			exitCode: 1,
			err:      errors.New("non-zero exit"),
		},
	}
	sut := &StatusAssertion{EqualsTo: &one}
	r := sut.verify(ctx)
	if !r.Success() {
		t.Errorf("expected success when exit matches expected non-zero value, got: %v", r.Errors())
	}
}

func TestStatusAssertion_CommandError_GreaterThan_Acceptable(t *testing.T) {
	one := 1
	ctx := planContext{
		commandResult: executionResult{
			exitCode: 2,
			err:      errors.New("non-zero exit"),
		},
	}
	sut := &StatusAssertion{GreaterThan: &one}
	r := sut.verify(ctx)
	if !r.Success() {
		t.Errorf("expected success: greater_than makes command errors acceptable, got: %v", r.Errors())
	}
}
