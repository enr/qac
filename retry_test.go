package qac

import (
	"strings"
	"testing"
)

// countingExecutor fails the first failFirst calls, then always succeeds.
type countingExecutor struct {
	callCount int
	failFirst int
}

func (e *countingExecutor) execute(_ Command) executionResult {
	e.callCount++
	if e.callCount <= e.failFirst {
		return executionResult{success: false, exitCode: 1}
	}
	return executionResult{success: true, exitCode: 0}
}

func retrySpec(retries int, retryDelay string) Spec {
	return Spec{
		Retries:    retries,
		RetryDelay: retryDelay,
		Command:    Command{Exe: "cmd"},
		Expectations: Expectations{
			StatusAssertion: StatusAssertion{EqualsTo: intPtr(0)},
		},
	}
}

// --- Core retry behaviour ---

func TestRetry_NoRetries_Success(t *testing.T) {
	e := &countingExecutor{failFirst: 0}
	sut := newLauncher(e)
	plan := TestPlan{Specs: map[string]Spec{"s": retrySpec(0, "")}}
	res := sut.Execute(plan)
	if errs := res.AllErrors(); len(errs) != 0 {
		t.Errorf("expected 0 errors, got: %v", errs)
	}
	if e.callCount != 1 {
		t.Errorf("expected exactly 1 call, got %d", e.callCount)
	}
}

func TestRetry_NoRetries_Failure(t *testing.T) {
	e := &countingExecutor{failFirst: 99}
	sut := newLauncher(e)
	plan := TestPlan{Specs: map[string]Spec{"s": retrySpec(0, "")}}
	res := sut.Execute(plan)
	if len(res.AllErrors()) == 0 {
		t.Error("expected errors for a failing spec with no retries")
	}
	if e.callCount != 1 {
		t.Errorf("expected exactly 1 call (no retry), got %d", e.callCount)
	}
}

func TestRetry_SucceedsOnFirstRetry(t *testing.T) {
	// fails once, passes on 2nd attempt (1 retry configured)
	e := &countingExecutor{failFirst: 1}
	sut := newLauncher(e)
	plan := TestPlan{Specs: map[string]Spec{"s": retrySpec(1, "")}}
	res := sut.Execute(plan)
	if errs := res.AllErrors(); len(errs) != 0 {
		t.Errorf("expected success after retry, got errors: %v", errs)
	}
	if e.callCount != 2 {
		t.Errorf("expected 2 total attempts, got %d", e.callCount)
	}
}

func TestRetry_TotalAttemptsIsRetriesPlusOne(t *testing.T) {
	// retries: 3 → 4 total attempts; fail all of them
	e := &countingExecutor{failFirst: 99}
	sut := newLauncher(e)
	plan := TestPlan{Specs: map[string]Spec{"s": retrySpec(3, "")}}
	sut.Execute(plan)
	if e.callCount != 4 {
		t.Errorf("expected 4 total attempts (retries+1), got %d", e.callCount)
	}
}

func TestRetry_AllAttemptsFailReportsError(t *testing.T) {
	e := &countingExecutor{failFirst: 99}
	sut := newLauncher(e)
	plan := TestPlan{Specs: map[string]Spec{"s": retrySpec(2, "")}}
	res := sut.Execute(plan)
	if len(res.AllErrors()) == 0 {
		t.Error("expected errors when all attempts fail")
	}
}

func TestRetry_PassesOnLastAllowedAttempt(t *testing.T) {
	// retries: 2 → 3 total; fail first 2, pass on 3rd
	e := &countingExecutor{failFirst: 2}
	sut := newLauncher(e)
	plan := TestPlan{Specs: map[string]Spec{"s": retrySpec(2, "")}}
	res := sut.Execute(plan)
	if errs := res.AllErrors(); len(errs) != 0 {
		t.Errorf("expected success on final retry, got: %v", errs)
	}
	if e.callCount != 3 {
		t.Errorf("expected 3 total calls, got %d", e.callCount)
	}
}

func TestRetry_IntermediateFailureInfoEntries(t *testing.T) {
	// fail once, succeed on retry — report should contain an info entry for the failed attempt
	e := &countingExecutor{failFirst: 1}
	sut := newLauncher(e)
	plan := TestPlan{Specs: map[string]Spec{"s": retrySpec(1, "")}}
	res := sut.Execute(plan)
	found := false
	for _, b := range res.Blocks() {
		for _, entry := range b.Entries() {
			if entry.Kind() == InfoType && strings.Contains(entry.Description(), "attempt 1") {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected an info entry describing the failed attempt 1")
	}
}

func TestRetry_WithDelay_Succeeds(t *testing.T) {
	// retry_delay: 1ms — just verify it doesn't break anything
	e := &countingExecutor{failFirst: 1}
	sut := newLauncher(e)
	plan := TestPlan{Specs: map[string]Spec{"s": retrySpec(1, "1ms")}}
	res := sut.Execute(plan)
	if errs := res.AllErrors(); len(errs) != 0 {
		t.Errorf("expected success with retry_delay, got: %v", errs)
	}
}

func TestRetry_InvalidDelay_ReportsConfigError(t *testing.T) {
	e := &countingExecutor{failFirst: 99}
	sut := newLauncher(e)
	plan := TestPlan{Specs: map[string]Spec{"s": retrySpec(1, "not-a-duration")}}
	res := sut.Execute(plan)
	// should have a config error about the invalid delay
	hasConfigErr := false
	for _, b := range res.Blocks() {
		for _, entry := range b.Entries() {
			if entry.Kind() == ErrorType {
				for _, err := range entry.Errors() {
					if strings.Contains(err.Error(), "retry_delay") {
						hasConfigErr = true
					}
				}
			}
		}
	}
	if !hasConfigErr {
		t.Error("expected a config error mentioning retry_delay for invalid duration")
	}
}

// --- YAML parsing ---

func TestRetryFieldsAccepted(t *testing.T) {
	input := `
specs:
  flaky:
    retries: 3
    retry_delay: 1s
    command:
      cli: my-tool
`
	plan, err := unmarshalPlan(t, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	spec := plan.Specs["flaky"]
	if spec.Retries != 3 {
		t.Errorf("retries = %d, want 3", spec.Retries)
	}
	if spec.RetryDelay != "1s" {
		t.Errorf("retry_delay = %q, want %q", spec.RetryDelay, "1s")
	}
}

func TestRetryTyposRejected(t *testing.T) {
	cases := []struct {
		name  string
		field string
		input string
	}{
		{
			name:  "retriies typo",
			field: "retriies",
			input: `
specs:
  test:
    retriies: 3
    command:
      cli: my-tool
`,
		},
		{
			name:  "retry_del typo",
			field: "retry_del",
			input: `
specs:
  test:
    retry_del: 1s
    command:
      cli: my-tool
`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := unmarshalPlan(t, tc.input)
			if err == nil {
				t.Fatalf("expected error for unknown field %q, got nil", tc.field)
			}
			if !strings.Contains(err.Error(), tc.field) {
				t.Errorf("error should mention %q, got: %v", tc.field, err)
			}
		})
	}
}

// --- Interaction with FailFast ---

func TestRetry_FailFast_TriggersAfterAllRetriesExhausted(t *testing.T) {
	// spec "a" fails all 2 attempts (retries:1); spec "b" should not run
	e := &countingExecutor{failFirst: 99}
	sut := newLauncher(e)
	plan := TestPlan{
		Specs: map[string]Spec{
			"a": retrySpec(1, ""),
			"b": retrySpec(0, ""),
		},
		specOrder: []string{"a", "b"},
	}
	sut.Execute(plan, FailFast())
	if e.callCount != 2 {
		t.Errorf("expected 2 calls (initial + 1 retry for a), got %d", e.callCount)
	}
}
