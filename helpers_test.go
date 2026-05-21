package qac

import "testing"

func TestBool_True(t *testing.T) {
	p := Bool(true)
	if p == nil || !*p {
		t.Errorf("Bool(true) = %v, want non-nil pointer to true", p)
	}
}

func TestBool_False(t *testing.T) {
	p := Bool(false)
	if p == nil || *p {
		t.Errorf("Bool(false) = %v, want non-nil pointer to false", p)
	}
}

func TestBool_Distinct(t *testing.T) {
	// Each call returns an independent pointer.
	a, b := Bool(true), Bool(true)
	if a == b {
		t.Error("Bool returns the same pointer on repeated calls")
	}
}

func TestInt_Value(t *testing.T) {
	p := Int(42)
	if p == nil || *p != 42 {
		t.Errorf("Int(42) = %v, want non-nil pointer to 42", p)
	}
}

func TestInt_Zero(t *testing.T) {
	p := Int(0)
	if p == nil || *p != 0 {
		t.Errorf("Int(0) = %v, want non-nil pointer to 0", p)
	}
}

func TestInt_Distinct(t *testing.T) {
	a, b := Int(1), Int(1)
	if a == b {
		t.Error("Int returns the same pointer on repeated calls")
	}
}

// Compile-time checks: verify the helpers satisfy the types expected by the
// public API fields they're designed for.
var _ *bool = Bool(true)
var _ *int = Int(0)

func TestBool_UsableInOutputAssertion(t *testing.T) {
	r := (&OutputAssertion{id: "stdout", IsEmpty: Bool(true)}).verify(outputCtx("", ""))
	if !r.Success() {
		t.Errorf("Bool(true) usable in IsEmpty: expected success for empty output, got: %v", r.Errors())
	}
}

func TestInt_UsableInStatusAssertion(t *testing.T) {
	ctx := planContext{commandResult: executionResult{exitCode: 0}}
	r := (&StatusAssertion{EqualsTo: Int(0)}).verify(ctx)
	if !r.Success() {
		t.Errorf("Int(0) usable in EqualsTo: expected success for exit code 0, got: %v", r.Errors())
	}
}
