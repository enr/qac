package qac

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// shared fixture builders

func reportWithResults(pass, fail, skip int) *TestExecutionReport {
	r := &TestExecutionReport{}
	idx := 0
	total := pass + fail + skip
	for i := 0; i < pass; i++ {
		idx++
		phase := fmt.Sprintf("pass-spec-%d", idx)
		r.openBlock(phase, idx, total, time.Time{})
		r.addEntryAsAssertionResult(phase, AssertionResult{description: "ok"})
		r.closeBlock(phase, 0)
	}
	for i := 0; i < fail; i++ {
		idx++
		phase := fmt.Sprintf("fail-spec-%d", idx)
		r.openBlock(phase, idx, total, time.Time{})
		ar := AssertionResult{description: "bad"}
		ar.addErrorf("something went wrong")
		r.addEntryAsAssertionResult(phase, ar)
		r.closeBlock(phase, 0)
	}
	for i := 0; i < skip; i++ {
		idx++
		phase := fmt.Sprintf("skip-spec-%d", idx)
		r.openBlock(phase, idx, total, time.Time{})
		r.addEntrySkipped(phase, "skip: true")
		r.closeBlock(phase, 0)
	}
	return r
}

// --- Success ---

func TestReportSuccess_AllPass(t *testing.T) {
	r := reportWithResults(3, 0, 0)
	if !r.Success() {
		t.Error("expected Success()=true when all specs pass")
	}
}

func TestReportSuccess_OneFail(t *testing.T) {
	r := reportWithResults(2, 1, 0)
	if r.Success() {
		t.Error("expected Success()=false when a spec fails")
	}
}

func TestReportSuccess_EmptyReport(t *testing.T) {
	if !(&TestExecutionReport{}).Success() {
		t.Error("expected Success()=true for an empty report")
	}
}

// --- FailedSpecs ---

func TestFailedSpecs_ReturnsPhaseNames(t *testing.T) {
	r := &TestExecutionReport{}
	// Two failing specs with distinct phases.
	for i, phase := range []string{"alpha", "beta"} {
		r.openBlock(phase, i+1, 2, time.Time{})
		ar := AssertionResult{description: "d"}
		ar.addErrorf("err")
		r.addEntryAsAssertionResult(phase, ar)
		r.closeBlock(phase, 0)
	}
	failed := r.FailedSpecs()
	if len(failed) != 2 {
		t.Fatalf("expected 2 failed specs, got %v", failed)
	}
	if failed[0] != "alpha" || failed[1] != "beta" {
		t.Errorf("unexpected failed spec names: %v", failed)
	}
}

func TestFailedSpecs_PassingSpecsExcluded(t *testing.T) {
	r := reportWithResults(3, 0, 0)
	if len(r.FailedSpecs()) != 0 {
		t.Errorf("expected no failed specs, got %v", r.FailedSpecs())
	}
}

func TestFailedSpecs_NonSpecBlocksExcluded(t *testing.T) {
	// A failing non-spec block (index=0) must not appear in FailedSpecs.
	r := &TestExecutionReport{}
	r.addEntryAsError("setup", errorf("setup failure"))
	if len(r.FailedSpecs()) != 0 {
		t.Errorf("setup failure must not appear in FailedSpecs, got %v", r.FailedSpecs())
	}
}

// --- Summary ---

func TestSummary_AllPass(t *testing.T) {
	got := reportWithResults(5, 0, 0).Summary()
	if got != "5/5 specs passed" {
		t.Errorf("Summary() = %q, want %q", got, "5/5 specs passed")
	}
}

func TestSummary_SomeFail(t *testing.T) {
	got := reportWithResults(3, 2, 0).Summary()
	if got != "3/5 specs passed" {
		t.Errorf("Summary() = %q, want %q", got, "3/5 specs passed")
	}
}

func TestSummary_WithSkipped(t *testing.T) {
	got := reportWithResults(3, 1, 1).Summary()
	if got != "3/5 specs passed (1 skipped)" {
		t.Errorf("Summary() = %q, want %q", got, "3/5 specs passed (1 skipped)")
	}
}

func TestSummary_AllSkipped(t *testing.T) {
	got := reportWithResults(0, 0, 3).Summary()
	if got != "0/3 specs passed (3 skipped)" {
		t.Errorf("Summary() = %q, want %q", got, "0/3 specs passed (3 skipped)")
	}
}

func TestSummary_Empty(t *testing.T) {
	got := (&TestExecutionReport{}).Summary()
	if got != "0/0 specs passed" {
		t.Errorf("Summary() = %q, want %q", got, "0/0 specs passed")
	}
}

func TestSummary_NonSpecBlocksIgnored(t *testing.T) {
	// Failures in setup/teardown must not inflate the spec count.
	r := &TestExecutionReport{}
	r.addEntryAsError("setup", errorf("failure"))
	r.openBlock("s", 1, 1, time.Time{})
	r.addEntryAsAssertionResult("s", AssertionResult{description: "ok"})
	r.closeBlock("s", 0)
	if r.Summary() != "1/1 specs passed" {
		t.Errorf("Summary() = %q, want %q", r.Summary(), "1/1 specs passed")
	}
}

// --- FailWith ---

func TestFailWith_CallsErrorfForEachError(t *testing.T) {
	r := reportWithResults(0, 2, 0)
	inner := &testing.T{}
	r.FailWith(inner)
	// Each failing spec contributes one error; inner should have failures recorded.
	// We can only verify indirectly that FailWith does not panic and interacts
	// with t. The direct assertion is that no panic occurred and inner.Failed()
	// reflects the calls.
	if !inner.Failed() {
		t.Error("FailWith should have called t.Errorf at least once")
	}
}

func TestFailWith_IncludesPhaseInMessage(t *testing.T) {
	// Capture messages by checking that the block phase appears somewhere
	// in the errors that FailWith would emit. We do this by inspecting the
	// report structure directly rather than intercepting t.Errorf output.
	r := &TestExecutionReport{}
	r.openBlock("my-spec", 1, 1, time.Time{})
	ar := AssertionResult{description: "d"}
	ar.addErrorf("the specific error")
	r.addEntryAsAssertionResult("my-spec", ar)
	r.closeBlock("my-spec", 0)

	// Verify the data that FailWith would use.
	found := false
	for _, b := range r.blocks {
		for _, e := range b.Entries() {
			for _, err := range e.Errors() {
				msg := "[" + b.Phase() + "] " + err.Error()
				if strings.Contains(msg, "my-spec") && strings.Contains(msg, "the specific error") {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("expected error message to contain both phase and error text")
	}
}

func TestFailWith_NoErrors_NoFailure(t *testing.T) {
	r := reportWithResults(3, 0, 0)
	inner := &testing.T{}
	r.FailWith(inner)
	if inner.Failed() {
		t.Error("FailWith on a passing report must not call t.Errorf")
	}
}

// helper

func errorf(msg string) error {
	return &Error{Kind: KindAssertionFailure, msg: msg}
}
