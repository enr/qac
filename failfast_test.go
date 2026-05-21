package qac

import "testing"

// --- FailFast unit/integration tests ---

func TestFailFast_StopsAfterFirstFailure(t *testing.T) {
	e := &trackingExecutor{
		results: map[string]executionResult{
			"fail": {success: false, exitCode: 1},
			"pass": {success: true, exitCode: 0},
		},
	}
	sut := newLauncher(e)
	plan := TestPlan{
		Specs: map[string]Spec{
			"a": {Command: Command{Exe: "fail"}, Expectations: Expectations{StatusAssertion: StatusAssertion{EqualsTo: intPtr(0)}}},
			"b": {Command: Command{Exe: "pass"}, Expectations: Expectations{StatusAssertion: StatusAssertion{EqualsTo: intPtr(0)}}},
			"c": {Command: Command{Exe: "pass"}, Expectations: Expectations{StatusAssertion: StatusAssertion{EqualsTo: intPtr(0)}}},
		},
		specOrder: []string{"a", "b", "c"},
	}
	sut.Execute(plan, FailFast())
	if e.wasExecuted("pass") {
		t.Error("expected b and c to not execute after a failed with FailFast")
	}
}

func TestFailFast_ReportContainsFailedBlock(t *testing.T) {
	e := &trackingExecutor{
		results: map[string]executionResult{
			"fail": {success: false, exitCode: 1},
		},
	}
	sut := newLauncher(e)
	plan := TestPlan{
		Specs: map[string]Spec{
			"a": {Command: Command{Exe: "fail"}, Expectations: Expectations{StatusAssertion: StatusAssertion{EqualsTo: intPtr(0)}}},
			"b": {Command: Command{Exe: "fail"}, Expectations: Expectations{StatusAssertion: StatusAssertion{EqualsTo: intPtr(0)}}},
		},
		specOrder: []string{"a", "b"},
	}
	res := sut.Execute(plan, FailFast())
	blocks := res.Blocks()
	if len(blocks) != 1 {
		t.Errorf("expected exactly 1 block in report (only a ran), got %d", len(blocks))
	}
	if !blocks[0].Failed() {
		t.Error("the single block should be marked as failed")
	}
}

func TestFailFast_NoFailure_AllSpecsRun(t *testing.T) {
	e := &trackingExecutor{}
	sut := newLauncher(e)
	plan := TestPlan{
		Specs: map[string]Spec{
			"a": {Command: Command{Exe: "pass"}},
			"b": {Command: Command{Exe: "pass"}},
		},
		specOrder: []string{"a", "b"},
	}
	sut.Execute(plan, FailFast())
	if !e.wasExecuted("pass") {
		t.Error("expected both specs to run when there is no failure")
	}
	if e.orderOf("pass") < 0 {
		t.Error("pass was expected to be executed")
	}
	// both a and b ran — orderOf returns the first occurrence; check count
	count := 0
	for _, cmd := range e.executed {
		if cmd == "pass" {
			count++
		}
	}
	if count != 2 {
		t.Errorf("expected pass to be executed twice, got %d", count)
	}
}

func TestFailFast_SkippedSpecDoesNotTrigger(t *testing.T) {
	e := &trackingExecutor{
		results: map[string]executionResult{
			"pass": {success: true, exitCode: 0},
		},
	}
	sut := newLauncher(e)
	plan := TestPlan{
		Specs: map[string]Spec{
			"a": {Tags: []string{"slow"}, Command: Command{Exe: "pass"}, Expectations: Expectations{StatusAssertion: StatusAssertion{EqualsTo: intPtr(0)}}},
			"b": {Tags: []string{"fast"}, Command: Command{Exe: "pass"}, Expectations: Expectations{StatusAssertion: StatusAssertion{EqualsTo: intPtr(0)}}},
		},
		specOrder: []string{"a", "b"},
	}
	// a is filtered (skipped), b should still run
	res := sut.Execute(plan, WithTags("fast"), FailFast())
	if errs := res.AllErrors(); len(errs) != 0 {
		t.Errorf("expected 0 errors, got: %v", errs)
	}
	if !e.wasExecuted("pass") {
		t.Error("expected b (fast tag) to run even though a was skipped")
	}
}

func TestFailFast_WithoutOption_AllSpecsRunDespiteFailure(t *testing.T) {
	e := &trackingExecutor{
		results: map[string]executionResult{
			"fail": {success: false, exitCode: 1},
			"pass": {success: true, exitCode: 0},
		},
	}
	sut := newLauncher(e)
	plan := TestPlan{
		Specs: map[string]Spec{
			"a": {Command: Command{Exe: "fail"}, Expectations: Expectations{StatusAssertion: StatusAssertion{EqualsTo: intPtr(0)}}},
			"b": {Command: Command{Exe: "pass"}, Expectations: Expectations{StatusAssertion: StatusAssertion{EqualsTo: intPtr(0)}}},
		},
		specOrder: []string{"a", "b"},
	}
	// no FailFast: b should still run even after a fails
	sut.Execute(plan)
	if !e.wasExecuted("pass") {
		t.Error("expected b to run even after a failed when FailFast is not set")
	}
}

// --- ReportBlock.Failed ---

func TestReportBlock_Failed_ErrorEntry(t *testing.T) {
	b := &ReportBlock{entries: []ReportEntry{{kind: ErrorType}}}
	if !b.Failed() {
		t.Error("expected Failed() to return true for a block with an error entry")
	}
}

func TestReportBlock_Failed_TimedOutEntry(t *testing.T) {
	b := &ReportBlock{entries: []ReportEntry{{kind: TimedOutType}}}
	if !b.Failed() {
		t.Error("expected Failed() to return true for a block with a timeout entry")
	}
}

func TestReportBlock_Failed_SuccessEntry(t *testing.T) {
	b := &ReportBlock{entries: []ReportEntry{{kind: SuccessType}}}
	if b.Failed() {
		t.Error("expected Failed() to return false for a block with only success entries")
	}
}

func TestReportBlock_Failed_EmptyBlock(t *testing.T) {
	b := &ReportBlock{}
	if b.Failed() {
		t.Error("expected Failed() to return false for an empty block")
	}
}
