package qac

import (
	"slices"
	"testing"
)

// trackingExecutor records the Cli (or Exe) of every command executed and
// returns configurable results keyed by that same string.
type trackingExecutor struct {
	executed []string
	results  map[string]executionResult
}

func (e *trackingExecutor) execute(c Command) executionResult {
	key := c.Cli
	if key == "" {
		key = c.Exe
	}
	e.executed = append(e.executed, key)
	if r, ok := e.results[key]; ok {
		return r
	}
	return executionResult{success: true, exitCode: 0}
}

func (e *trackingExecutor) wasExecuted(cmd string) bool {
	return slices.Contains(e.executed, cmd)
}

// orderOf returns the 0-based position of cmd in executed, or -1 if absent.
func (e *trackingExecutor) orderOf(cmd string) int {
	for i, k := range e.executed {
		if k == cmd {
			return i
		}
	}
	return -1
}

// --- spec-level setup ---

func TestSpecSetup_RunsBeforeCommand(t *testing.T) {
	tr := &trackingExecutor{}
	sut := newLauncher(tr)
	zero := 0
	spec := Spec{
		Setup:   []Command{{Cli: "setup-cmd"}},
		Command: Command{Cli: "main-cmd"},
		Expectations: Expectations{
			StatusAssertion: StatusAssertion{EqualsTo: &zero},
		},
	}
	sut.Execute(TestPlan{Specs: map[string]Spec{"s": spec}})

	if !tr.wasExecuted("setup-cmd") {
		t.Error("setup command was not executed")
	}
	if !tr.wasExecuted("main-cmd") {
		t.Error("main command was not executed")
	}
	if tr.orderOf("setup-cmd") >= tr.orderOf("main-cmd") {
		t.Error("setup command did not run before the main command")
	}
}

func TestSpecSetup_Failure_SkipsCommand(t *testing.T) {
	tr := &trackingExecutor{
		results: map[string]executionResult{
			"setup-cmd": {success: false, exitCode: 1},
		},
	}
	sut := newLauncher(tr)
	sut.Execute(TestPlan{Specs: map[string]Spec{"s": {
		Setup:   []Command{{Cli: "setup-cmd"}},
		Command: Command{Cli: "main-cmd"},
	}}})

	if tr.wasExecuted("main-cmd") {
		t.Error("main command should not run when setup fails")
	}
}

// --- spec-level teardown ---

func TestSpecTeardown_RunsAfterSuccess(t *testing.T) {
	tr := &trackingExecutor{}
	sut := newLauncher(tr)
	zero := 0
	spec := Spec{
		Teardown: []Command{{Cli: "teardown-cmd"}},
		Command:  Command{Cli: "main-cmd"},
		Expectations: Expectations{
			StatusAssertion: StatusAssertion{EqualsTo: &zero},
		},
	}
	sut.Execute(TestPlan{Specs: map[string]Spec{"s": spec}})

	if !tr.wasExecuted("teardown-cmd") {
		t.Error("teardown command was not executed after a passing spec")
	}
	if tr.orderOf("main-cmd") >= tr.orderOf("teardown-cmd") {
		t.Error("teardown command did not run after the main command")
	}
}

func TestSpecTeardown_RunsAfterCommandFailure(t *testing.T) {
	tr := &trackingExecutor{
		results: map[string]executionResult{
			"main-cmd": {success: false, exitCode: 1},
		},
	}
	sut := newLauncher(tr)
	sut.Execute(TestPlan{Specs: map[string]Spec{"s": {
		Teardown: []Command{{Cli: "teardown-cmd"}},
		Command:  Command{Cli: "main-cmd"},
		Expectations: Expectations{
			StatusAssertion: StatusAssertion{EqualsTo: intPtr(0)},
		},
	}}})

	if !tr.wasExecuted("teardown-cmd") {
		t.Error("teardown must run even when the command fails")
	}
}

func TestSpecTeardown_RunsAfterSetupFailure(t *testing.T) {
	tr := &trackingExecutor{
		results: map[string]executionResult{
			"setup-cmd": {success: false, exitCode: 1},
		},
	}
	sut := newLauncher(tr)
	sut.Execute(TestPlan{Specs: map[string]Spec{"s": {
		Setup:    []Command{{Cli: "setup-cmd"}},
		Teardown: []Command{{Cli: "teardown-cmd"}},
		Command:  Command{Cli: "main-cmd"},
	}}})

	if !tr.wasExecuted("teardown-cmd") {
		t.Error("teardown must run even when setup fails")
	}
	if tr.wasExecuted("main-cmd") {
		t.Error("main command must not run when setup fails")
	}
}

func TestSpecTeardown_NotRunWhenSkipped(t *testing.T) {
	tr := &trackingExecutor{}
	sut := newLauncher(tr)
	sut.Execute(TestPlan{Specs: map[string]Spec{"s": {
		Skip:     true,
		Teardown: []Command{{Cli: "teardown-cmd"}},
		Command:  Command{Cli: "main-cmd"},
	}}})

	if tr.wasExecuted("teardown-cmd") {
		t.Error("teardown must not run when the spec is skipped")
	}
}

func TestSpecTeardown_AllCommandsRunOnFailure(t *testing.T) {
	// First teardown command fails; second should still run.
	tr := &trackingExecutor{
		results: map[string]executionResult{
			"teardown-1": {success: false, exitCode: 1},
		},
	}
	sut := newLauncher(tr)
	sut.Execute(TestPlan{Specs: map[string]Spec{"s": {
		Teardown: []Command{{Cli: "teardown-1"}, {Cli: "teardown-2"}},
		Command:  Command{Cli: "main-cmd"},
	}}})

	if !tr.wasExecuted("teardown-2") {
		t.Error("second teardown command should run even when first teardown fails")
	}
}

// --- plan-level setup ---

func TestPlanSetup_RunsBeforeSpecs(t *testing.T) {
	tr := &trackingExecutor{}
	sut := newLauncher(tr)
	plan := TestPlan{
		Setup: []Command{{Cli: "plan-setup"}},
		Specs: map[string]Spec{"s": {Command: Command{Cli: "spec-cmd"}}},
	}
	sut.Execute(plan)

	if !tr.wasExecuted("plan-setup") {
		t.Error("plan setup was not executed")
	}
	if tr.orderOf("plan-setup") >= tr.orderOf("spec-cmd") {
		t.Error("plan setup did not run before spec commands")
	}
}

func TestPlanSetupFailure_SkipsAllSpecs(t *testing.T) {
	tr := &trackingExecutor{
		results: map[string]executionResult{
			"plan-setup": {success: false, exitCode: 1},
		},
	}
	sut := newLauncher(tr)
	plan := TestPlan{
		Setup: []Command{{Cli: "plan-setup"}},
		Specs: map[string]Spec{
			"s1": {Command: Command{Cli: "spec-1"}},
			"s2": {Command: Command{Cli: "spec-2"}},
		},
	}
	sut.Execute(plan)

	if tr.wasExecuted("spec-1") || tr.wasExecuted("spec-2") {
		t.Error("specs must not run when plan setup fails")
	}
}

func TestPlanSetupFailure_TeardownStillRuns(t *testing.T) {
	tr := &trackingExecutor{
		results: map[string]executionResult{
			"plan-setup": {success: false, exitCode: 1},
		},
	}
	sut := newLauncher(tr)
	plan := TestPlan{
		Setup:    []Command{{Cli: "plan-setup"}},
		Teardown: []Command{{Cli: "plan-teardown"}},
		Specs:    map[string]Spec{"s": {Command: Command{Cli: "spec-cmd"}}},
	}
	sut.Execute(plan)

	if !tr.wasExecuted("plan-teardown") {
		t.Error("plan teardown must run even when plan setup fails")
	}
}

// --- plan-level teardown ---

func TestPlanTeardown_RunsAfterAllSpecs(t *testing.T) {
	tr := &trackingExecutor{}
	sut := newLauncher(tr)
	plan := TestPlan{
		Teardown: []Command{{Cli: "plan-teardown"}},
		Specs: map[string]Spec{
			"s1": {Command: Command{Cli: "spec-1"}},
			"s2": {Command: Command{Cli: "spec-2"}},
		},
		specOrder: []string{"s1", "s2"},
	}
	sut.Execute(plan)

	if !tr.wasExecuted("plan-teardown") {
		t.Error("plan teardown was not executed")
	}
	// Teardown must run after both specs.
	if tr.orderOf("plan-teardown") <= tr.orderOf("spec-1") ||
		tr.orderOf("plan-teardown") <= tr.orderOf("spec-2") {
		t.Error("plan teardown did not run after all specs")
	}
}

func TestPlanTeardown_RunsEvenWhenSpecFails(t *testing.T) {
	tr := &trackingExecutor{
		results: map[string]executionResult{
			"spec-cmd": {success: false, exitCode: 1},
		},
	}
	sut := newLauncher(tr)
	plan := TestPlan{
		Teardown: []Command{{Cli: "plan-teardown"}},
		Specs:    map[string]Spec{"s": {Command: Command{Cli: "spec-cmd"}}},
	}
	sut.Execute(plan)

	if !tr.wasExecuted("plan-teardown") {
		t.Error("plan teardown must run even when a spec fails")
	}
}

// intPtr is a helper for pointer-to-int literals.
func intPtr(v int) *int { return &v }
