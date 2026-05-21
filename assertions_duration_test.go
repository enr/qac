package qac

import (
	"testing"
	"time"
)

func durationCtx(d time.Duration) planContext {
	return planContext{
		commandResult: executionResult{
			execution: "test-cmd",
			duration:  d,
		},
	}
}

// --- Max ---

func TestDurationMax_Pass(t *testing.T) {
	r := (&DurationAssertion{Max: "2s"}).verify(durationCtx(500 * time.Millisecond))
	if !r.Success() {
		t.Errorf("expected success when duration < max, got: %v", r.Errors())
	}
}

func TestDurationMax_ExactlyAtLimit_Pass(t *testing.T) {
	r := (&DurationAssertion{Max: "1s"}).verify(durationCtx(time.Second))
	if !r.Success() {
		t.Errorf("expected success when duration == max, got: %v", r.Errors())
	}
}

func TestDurationMax_Fail(t *testing.T) {
	r := (&DurationAssertion{Max: "500ms"}).verify(durationCtx(time.Second))
	if r.Success() {
		t.Error("expected failure when duration > max")
	}
}

func TestDurationMax_InvalidDuration(t *testing.T) {
	r := (&DurationAssertion{Max: "notaduration"}).verify(durationCtx(time.Second))
	if r.Success() {
		t.Error("expected failure for invalid max duration string")
	}
}

// --- Min ---

func TestDurationMin_Pass(t *testing.T) {
	r := (&DurationAssertion{Min: "100ms"}).verify(durationCtx(500 * time.Millisecond))
	if !r.Success() {
		t.Errorf("expected success when duration > min, got: %v", r.Errors())
	}
}

func TestDurationMin_ExactlyAtLimit_Pass(t *testing.T) {
	r := (&DurationAssertion{Min: "200ms"}).verify(durationCtx(200 * time.Millisecond))
	if !r.Success() {
		t.Errorf("expected success when duration == min, got: %v", r.Errors())
	}
}

func TestDurationMin_Fail(t *testing.T) {
	r := (&DurationAssertion{Min: "1s"}).verify(durationCtx(10 * time.Millisecond))
	if r.Success() {
		t.Error("expected failure when duration < min")
	}
}

func TestDurationMin_InvalidDuration(t *testing.T) {
	r := (&DurationAssertion{Min: "2 seconds"}).verify(durationCtx(time.Second))
	if r.Success() {
		t.Error("expected failure for invalid min duration string")
	}
}

// --- Both bounds ---

func TestDurationBoth_Pass(t *testing.T) {
	r := (&DurationAssertion{Min: "100ms", Max: "2s"}).verify(durationCtx(500 * time.Millisecond))
	if !r.Success() {
		t.Errorf("expected success when duration is within [min, max], got: %v", r.Errors())
	}
}

func TestDurationBoth_BelowMin(t *testing.T) {
	r := (&DurationAssertion{Min: "100ms", Max: "2s"}).verify(durationCtx(10 * time.Millisecond))
	if r.Success() {
		t.Error("expected failure when duration < min")
	}
	if len(r.Errors()) != 1 {
		t.Errorf("expected exactly 1 error, got %d: %v", len(r.Errors()), r.Errors())
	}
}

func TestDurationBoth_ExceedsMax(t *testing.T) {
	r := (&DurationAssertion{Min: "100ms", Max: "500ms"}).verify(durationCtx(time.Second))
	if r.Success() {
		t.Error("expected failure when duration > max")
	}
	if len(r.Errors()) != 1 {
		t.Errorf("expected exactly 1 error, got %d: %v", len(r.Errors()), r.Errors())
	}
}

func TestDurationBoth_BothInvalid(t *testing.T) {
	r := (&DurationAssertion{Min: "bad", Max: "also-bad"}).verify(durationCtx(time.Second))
	if r.Success() {
		t.Error("expected failure for two invalid duration strings")
	}
	if len(r.Errors()) != 2 {
		t.Errorf("expected 2 errors (one per invalid field), got %d: %v", len(r.Errors()), r.Errors())
	}
}

// --- Launcher integration: duration is measured and the assertion runs ---

func TestLauncher_DurationMax_Pass(t *testing.T) {
	e := &fixedValueExecutor{success: true, exitCode: 0}
	sut := newLauncher(e)
	zero := 0
	spec := Spec{
		Command: Command{Exe: "true"},
		Expectations: Expectations{
			StatusAssertion:   StatusAssertion{EqualsTo: &zero},
			DurationAssertion: DurationAssertion{Max: "10s"},
		},
	}
	res := sut.Execute(TestPlan{Specs: map[string]Spec{"s": spec}})
	if errs := res.AllErrors(); len(errs) != 0 {
		t.Errorf("expected 0 errors, got: %v", errs)
	}
}

func TestLauncher_DurationMax_Fail(t *testing.T) {
	e := &fixedValueExecutor{success: true, exitCode: 0}
	sut := newLauncher(e)
	zero := 0
	spec := Spec{
		Command: Command{Exe: "true"},
		Expectations: Expectations{
			StatusAssertion:   StatusAssertion{EqualsTo: &zero},
			DurationAssertion: DurationAssertion{Max: "0s"}, // always fails
		},
	}
	res := sut.Execute(TestPlan{Specs: map[string]Spec{"s": spec}})
	if len(res.AllErrors()) == 0 {
		t.Error("expected at least one duration error")
	}
}

func TestLauncher_DurationNotVerifiedWhenEmpty(t *testing.T) {
	e := &fixedValueExecutor{success: true, exitCode: 0}
	sut := newLauncher(e)
	zero := 0
	spec := Spec{
		Command: Command{Exe: "true"},
		Expectations: Expectations{
			StatusAssertion: StatusAssertion{EqualsTo: &zero},
			// DurationAssertion is zero-value: no check should run
		},
	}
	res := sut.Execute(TestPlan{Specs: map[string]Spec{"s": spec}})
	if errs := res.AllErrors(); len(errs) != 0 {
		t.Errorf("expected 0 errors with no duration assertion, got: %v", errs)
	}
}
